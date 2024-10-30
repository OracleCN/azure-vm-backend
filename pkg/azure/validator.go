package azure

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// Credentials 包含从 az ad sp create-for-rbac 获取的凭据信息
type Credentials struct {
	TenantID     string // tenant
	ClientID     string // appId
	ClientSecret string // password
	DisplayName  string // displayName
}

// ValidationResult 包含验证结果的详细信息
type ValidationResult struct {
	Valid       bool
	Message     string
	ValidatedAt time.Time
	Error       error
}

// Validator Azure 凭据验证器
type Validator struct {
	timeout time.Duration
}

// NewValidator 创建一个新的验证器实例
func NewValidator(timeout time.Duration) *Validator {
	if timeout == 0 {
		timeout = 60 * time.Second // 默认超时时间
	}
	return &Validator{
		timeout: timeout,
	}
}

// ValidateWithContext 验证服务主体凭据
func (v *Validator) ValidateWithContext(ctx context.Context, credentials Credentials) ValidationResult {
	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	result := ValidationResult{
		ValidatedAt: time.Now(),
	}

	// 1. 并行执行格式验证
	formatErrChan := make(chan error, 1)
	go func() {
		formatErrChan <- validateCredentialsFormat(credentials)
	}()

	// 等待格式验证结果
	select {
	case err := <-formatErrChan:
		if err != nil {
			result.Error = fmt.Errorf("凭据格式验证失败: %w", err)
			result.Message = "Azure 凭据格式无效"
			return result
		}
	case <-ctx.Done():
		result.Error = fmt.Errorf("凭据格式验证超时")
		result.Message = "验证超时"
		return result
	}

	// 2. 创建凭据对象
	cred, err := azidentity.NewClientSecretCredential(
		credentials.TenantID,
		credentials.ClientID,
		credentials.ClientSecret,
		nil,
	)
	if err != nil {
		result.Error = fmt.Errorf("创建凭据对象失败: %w", err)
		result.Message = "创建 Azure 凭据失败"
		return result
	}

	// 3. 并行执行认证检查
	var wg sync.WaitGroup
	type validationTask struct {
		name string
		err  error
	}

	taskChan := make(chan validationTask, 2) // 缓冲通道用于存储任务结果

	// 启动订阅检查协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		subscriptionID, err := getSubscriptionID(ctx, cred)
		if err != nil {
			taskChan <- validationTask{"subscription", err}
			return
		}

		// 创建资源管理客户端并验证权限
		clientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
		if err != nil {
			taskChan <- validationTask{"client", err}
			return
		}

		client := clientFactory.NewResourceGroupsClient()
		pager := client.NewListPager(nil)
		_, err = pager.NextPage(ctx)
		taskChan <- validationTask{"permission", err}
	}()

	// 等待所有验证任务完成或上下文取消
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(taskChan)
	}()

	select {
	case <-done:
		// 检查任务结果
		for task := range taskChan {
			if task.err != nil {
				switch task.name {
				case "subscription":
					result.Error = fmt.Errorf("获取订阅失败: %w", task.err)
					result.Message = "无法获取 Azure 订阅"
				case "client":
					result.Error = fmt.Errorf("创建客户端失败: %w", task.err)
					result.Message = "初始化 Azure 客户端失败"
				case "permission":
					result.Error = fmt.Errorf("权限验证失败: %w", task.err)
					result.Message = "Azure 权限验证失败"
				}
				return result
			}
		}
	case <-ctx.Done():
		result.Error = fmt.Errorf("验证超时: %w", ctx.Err())
		result.Message = "Azure 验证超时"
		return result
	}

	result.Valid = true
	result.Message = "成功验证 Azure 凭据"
	return result
}

// getSubscriptionID 获取服务主体可访问的第一个订阅 ID
func getSubscriptionID(ctx context.Context, cred *azidentity.ClientSecretCredential) (string, error) {
	// 创建订阅客户端
	subsClient, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return "", fmt.Errorf("创建订阅客户端失败: %w", err)
	}

	// 获取所有订阅
	pager := subsClient.NewListPager(nil)
	page, err := pager.NextPage(ctx)
	if err != nil {
		return "", fmt.Errorf("获取订阅列表失败: %w", err)
	}

	// 检查是否有可用的订阅
	if len(page.Value) == 0 {
		return "", fmt.Errorf("未找到可用的订阅，请确保服务主体有正确的访问权限")
	}

	// 返回第一个可用的订阅 ID
	return *page.Value[0].SubscriptionID, nil
}

// validateCredentialsFormat 验证凭据格式
func validateCredentialsFormat(creds Credentials) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 4) // 缓冲通道存储验证错误

	// 并行验证各个字段
	wg.Add(4)

	// 验证 TenantID
	go func() {
		defer wg.Done()
		if err := validateUUID("tenant ID", creds.TenantID); err != nil {
			errChan <- err
		}
	}()

	// 验证 ClientID
	go func() {
		defer wg.Done()
		if err := validateUUID("client ID (App ID)", creds.ClientID); err != nil {
			errChan <- err
		}
	}()

	// 验证 ClientSecret
	go func() {
		defer wg.Done()
		if err := validateClientSecret(creds.ClientSecret); err != nil {
			errChan <- err
		}
	}()

	// 验证 DisplayName
	go func() {
		defer wg.Done()
		if err := validateDisplayName(creds.DisplayName); err != nil {
			errChan <- err
		}
	}()

	// 等待所有验证完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errors []string
	for err := range errChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("凭据格式验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateUUID 验证 UUID 格式
func validateUUID(fieldName, uuid string) error {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return fmt.Errorf("%s 不能为空", fieldName)
	}

	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidPattern.MatchString(uuid) {
		return fmt.Errorf("%s 格式无效: %s", fieldName, uuid)
	}

	return nil
}

// validateClientSecret 验证 Client Secret 格式
func validateClientSecret(secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return fmt.Errorf("client secret (password) 不能为空")
	}
	return nil
}

// validateDisplayName 验证显示名称
func validateDisplayName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("display name 不能为空")
	}
	return nil
}
