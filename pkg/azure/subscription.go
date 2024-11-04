package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"go.uber.org/zap"
)

// SpendingLimit 消费限制枚举类型
type SpendingLimit string

const (
	SpendingLimitOn  SpendingLimit = "On"
	SpendingLimitOff SpendingLimit = "Off"
)

// SubscriptionDetail 订阅详细信息
type SubscriptionDetail struct {
	SubscriptionID       string                 // 订阅ID
	DisplayName          string                 // 显示名称
	State                string                 // 状态
	SubscriptionPolicies map[string]interface{} // 订阅策略
	AuthorizationSource  string                 // 授权来源
	FetchedAt            time.Time              // 数据获取时间
	StartDate            *time.Time             // 订阅开始时间
	EndDate              *time.Time             // 订阅结束时间
	SpendingLimit        string                 // 消费限制
	SubscriptionType     string                 // 订阅类型
}

// Fetcher Azure数据获取器
type Fetcher struct {
	credentials *Credentials
	logger      *zap.Logger
	timeout     time.Duration
	mutex       sync.RWMutex
}

// NewFetcher 创建新的Azure数据获取器实例
func NewFetcher(credentials *Credentials, logger *zap.Logger, timeout time.Duration) *Fetcher {
	if timeout == 0 {
		timeout = 60 * time.Second // 默认超时时间
	}
	return &Fetcher{
		credentials: credentials,
		logger:      logger,
		timeout:     timeout,
	}
}

// normalizeSpendingLimit 规范化消费限制值
func normalizeSpendingLimit(value string) string {
	switch strings.ToLower(value) {
	case "on", "1", "true":
		return string(SpendingLimitOn)
	case "off", "0", "false":
		return string(SpendingLimitOff)
	default:
		return value
	}
}

// FetchSubscriptionDetails 获取订阅详细信息
func (f *Fetcher) FetchSubscriptionDetails(ctx context.Context) ([]SubscriptionDetail, error) {
	f.logger.Info("开始获取Azure订阅详细信息")
	startTime := time.Now()

	// 创建上下文
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	// 创建凭据对象
	cred, err := azidentity.NewClientSecretCredential(
		f.credentials.TenantID,
		f.credentials.ClientID,
		f.credentials.ClientSecret,
		nil,
	)
	if err != nil {
		f.logger.Error("创建Azure凭据失败",
			zap.Error(err),
			zap.String("tenantId", f.credentials.TenantID),
		)
		return nil, fmt.Errorf("创建Azure凭据失败: %w", err)
	}

	// 创建订阅客户端
	client, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		f.logger.Error("创建订阅客户端失败", zap.Error(err))
		return nil, fmt.Errorf("创建订阅客户端失败: %w", err)
	}

	// 获取所有订阅
	pager := client.NewListPager(nil)
	var subscriptions []SubscriptionDetail
	subChan := make(chan SubscriptionDetail, 10) // 使用带缓冲的channel
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// 启动错误收集的goroutine
	var collectedErrors []error
	var errWg sync.WaitGroup
	errWg.Add(1)
	go func() {
		defer errWg.Done()
		for err := range errChan {
			f.mutex.Lock()
			collectedErrors = append(collectedErrors, err)
			f.mutex.Unlock()
		}
	}()

	// 启动收集结果的goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for sub := range subChan {
			f.mutex.Lock()
			subscriptions = append(subscriptions, sub)
			f.mutex.Unlock()
		}
	}()

	var subWg sync.WaitGroup
	hasResults := false

	// 遍历订阅列表
	for {
		page, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) {
				f.logger.Error("Azure API返回错误",
					zap.Int("StatusCode", respErr.StatusCode),
					zap.String("ErrorCode", respErr.ErrorCode),
					zap.Error(err),
				)
				return nil, fmt.Errorf("azure API错误: %w", err)
			}

			if !hasResults {
				f.logger.Error("获取订阅列表失败", zap.Error(err))
				return nil, fmt.Errorf("获取订阅列表失败: %w", err)
			}
			break
		}

		if len(page.Value) == 0 {
			break
		}

		hasResults = true
		// 并发处理每个订阅的详细信息
		for _, sub := range page.Value {
			subWg.Add(1)
			go func(subscription *armsubscription.Subscription) {
				defer subWg.Done()

				detail := SubscriptionDetail{
					FetchedAt: time.Now(),
				}

				// 处理基本信息
				if subscription.ID != nil {
					detail.SubscriptionID = extractSubscriptionID(*subscription.ID)
				}
				if subscription.DisplayName != nil {
					detail.DisplayName = *subscription.DisplayName
				}
				if subscription.State != nil {
					detail.State = string(*subscription.State)
				}
				if subscription.AuthorizationSource != nil {
					detail.AuthorizationSource = *subscription.AuthorizationSource
				}

				// 处理订阅策略
				if subscription.SubscriptionPolicies != nil {
					detail.SubscriptionPolicies = make(map[string]interface{})
					if subscription.SubscriptionPolicies.LocationPlacementID != nil {
						detail.SubscriptionPolicies["locationPlacementId"] = *subscription.SubscriptionPolicies.LocationPlacementID
					}
					if subscription.SubscriptionPolicies.QuotaID != nil {
						detail.SubscriptionPolicies["quotaId"] = *subscription.SubscriptionPolicies.QuotaID
						quotaID := *subscription.SubscriptionPolicies.QuotaID

						// 根据 QuotaID 判断订阅类型
						switch {
						case strings.HasPrefix(quotaID, "AzureForStudents"):
							detail.SubscriptionType = "Student"
							startDate := time.Now().AddDate(0, -12, 0)
							endDate := startDate.AddDate(1, 0, 0)
							detail.StartDate = &startDate
							detail.EndDate = &endDate

						case strings.Contains(quotaID, "FreeTrial"):
							detail.SubscriptionType = "FreeTrial"
							startDate := time.Now().AddDate(0, 0, -30)
							endDate := startDate.AddDate(0, 0, 30)
							detail.StartDate = &startDate
							detail.EndDate = &endDate

						case strings.Contains(quotaID, "PayAsYouGo"):
							detail.SubscriptionType = "PayAsYouGo"
							startDate := time.Now()
							detail.StartDate = &startDate

						case strings.Contains(quotaID, "MSDN"):
							detail.SubscriptionType = "MSDN"
							startDate := time.Now().AddDate(0, -1, 0)
							detail.StartDate = &startDate

						default:
							detail.SubscriptionType = "Other"
							f.logger.Info("未识别的订阅类型",
								zap.String("quotaId", quotaID),
								zap.String("subscriptionId", detail.SubscriptionID),
							)
						}
					}
					if subscription.SubscriptionPolicies.SpendingLimit != nil {
						spendingLimit := string(*subscription.SubscriptionPolicies.SpendingLimit)
						detail.SpendingLimit = normalizeSpendingLimit(spendingLimit)
						detail.SubscriptionPolicies["spendingLimit"] = detail.SpendingLimit
					}
				}

				f.logger.Info("订阅信息",
					zap.String("subscriptionId", detail.SubscriptionID),
					zap.String("type", detail.SubscriptionType),
					zap.Any("startDate", detail.StartDate),
					zap.Any("endDate", detail.EndDate),
					zap.String("spendingLimit", detail.SpendingLimit),
				)

				subChan <- detail
			}(sub)
		}
	}

	// 等待所有处理goroutine完成
	subWg.Wait()
	// 关闭channel
	close(subChan)
	close(errChan)
	// 等待收集结果的goroutine完成
	wg.Wait()

	if len(collectedErrors) > 0 {
		f.logger.Warn("处理订阅信息时发生错误",
			zap.Int("错误数量", len(collectedErrors)),
			zap.Errors("errors", collectedErrors),
		)
	}
	if len(subscriptions) > 0 {
		f.logger.Info("Azure订阅信息获取完成",
			zap.Int("订阅数量", len(subscriptions)),
			zap.Duration("耗时", time.Since(startTime)),
		)
	} else {
		f.logger.Warn("未找到任何订阅信息",
			zap.Duration("耗时", time.Since(startTime)),
		)
		return nil, fmt.Errorf("未找到任何订阅")
	}

	return subscriptions, nil
}
