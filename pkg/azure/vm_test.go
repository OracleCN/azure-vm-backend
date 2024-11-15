package azure

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"log"
	"testing"
	"time"
)

func TestSetVMDNSLabelManual(t *testing.T) {
	// 1. 准备凭据
	credentials := &Credentials{
		TenantID:     "d6fcb345-610d-4ed2-ac0a-cfa940f9fa5f",     // 替换为你的租户ID
		ClientID:     "d91bca22-19fe-4c8a-82f1-c0bcd2cdbb94",     // 替换为你的客户端ID
		ClientSecret: "9t.8Q~I3nIFtGxr4elqlEPB2TXdfDarrwSRbpaXk", // 替换为你的客户端密钥
	}

	// 2. 创建logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("创建logger失败: %v", err)
	}
	defer logger.Sync()

	// 3. 创建VMFetcher
	fetcher := NewVMFetcher(credentials, logger, 30*time.Second)

	ctx := context.Background()

	fqdn, err := fetcher.SetVMDNSLabel(
		ctx,
		"1802f3c5-7ca4-4109-b437-3e34932602c9", // 替换为你的订阅ID
		"TEST",                                 // 替换为你的资源组名称
		"test-ip",                              // 替换为你的公共IP名称
		"test-dns-label-1111",                  // 你想设置的DNS标签
	)

	// 5. 检查结果
	if err != nil {
		log.Fatalf("设置DNS标签失败: %v", err)
	}

	fmt.Printf("DNS标签设置成功！\nFQDN: %s\n", fqdn)
}

// TestVMOperations 测试VM操作的集成测试
func TestVMOperations(t *testing.T) {
	// 从环境变量获取Azure凭据
	creds := &Credentials{
		TenantID:     "d6fcb345-610d-4ed2-ac0a-cfa940f9fa5f",     // 替换为你的租户ID
		ClientID:     "d91bca22-19fe-4c8a-82f1-c0bcd2cdbb94",     // 替换为你的客户端ID
		ClientSecret: "9t.8Q~I3nIFtGxr4elqlEPB2TXdfDarrwSRbpaXk", // 替换为你的客户端密钥
	}

	// 测试配置
	testConfig := struct {
		subscriptionID string
		resourceGroup  string
		vmName         string
	}{
		subscriptionID: "1802f3c5-7ca4-4109-b437-3e34932602c9",
		resourceGroup:  "TEST",
		vmName:         "test",
	}

	if testConfig.subscriptionID == "" || testConfig.resourceGroup == "" || testConfig.vmName == "" {
		t.Skip("需要设置AZURE_SUBSCRIPTION_ID, AZURE_RESOURCE_GROUP, AZURE_VM_NAME环境变量")
	}

	// 创建logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("创建logger失败: %v", err)
	}
	defer logger.Sync()

	// 创建context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 创建VMFetcher实例
	fetcher := NewVMFetcher(creds, logger, 5*time.Minute)

	// 定义测试用的VM详情
	vmDetails := VMDetails{
		SubscriptionID: testConfig.subscriptionID,
		ResourceGroup:  testConfig.resourceGroup,
		Name:           testConfig.vmName,
	}

	t.Run("测试获取VM状态", func(t *testing.T) {
		status, err := fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
		assert.NoError(t, err)
		t.Logf("当前VM状态: %s", status)
	})

	t.Run("测试完整的VM操作流程", func(t *testing.T) {
		// 1. 确保VM处于运行状态（如果需要的话启动VM）
		initialStatus, err := fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
		assert.NoError(t, err)

		if initialStatus != "running" {
			t.Logf("VM当前状态为%s，尝试启动VM", initialStatus)
			err = fetcher.VMOperation(ctx, VMOperationStart, vmDetails, nil)
			assert.NoError(t, err)

			// 等待VM启动
			time.Sleep(30 * time.Second)

			status, err := fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
			assert.NoError(t, err)
			assert.Equal(t, "running", status, "VM应该处于运行状态")
		}

		// 2. 测试停止VM（正常关机）
		t.Log("测试正常关机...")
		err = fetcher.VMOperation(ctx, VMOperationStop, vmDetails, &OperationOptions{Force: false})
		assert.NoError(t, err)

		// 等待VM停止
		time.Sleep(30 * time.Second)

		status, err := fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
		assert.NoError(t, err)
		assert.Equal(t, "stopped", status, "VM应该处于停止状态")

		// 3. 测试启动VM
		t.Log("测试启动VM...")
		err = fetcher.VMOperation(ctx, VMOperationStart, vmDetails, nil)
		assert.NoError(t, err)

		// 等待VM启动
		time.Sleep(30 * time.Second)

		status, err = fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
		assert.NoError(t, err)
		assert.Equal(t, "running", status, "VM应该处于运行状态")

		// 4. 测试重启VM
		t.Log("测试重启VM...")
		err = fetcher.VMOperation(ctx, VMOperationRestart, vmDetails, nil)
		assert.NoError(t, err)

		// 等待重启完成
		time.Sleep(60 * time.Second)

		status, err = fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
		assert.NoError(t, err)
		assert.Equal(t, "running", status, "重启后VM应该处于运行状态")

		// 5. 测试强制停止（释放资源）
		t.Log("测试强制停止VM...")
		err = fetcher.VMOperation(ctx, VMOperationStop, vmDetails, &OperationOptions{Force: true})
		assert.NoError(t, err)

		// 等待操作完成
		time.Sleep(30 * time.Second)

		status, err = fetcher.GetVMStatus(ctx, testConfig.subscriptionID, testConfig.resourceGroup, testConfig.vmName)
		assert.NoError(t, err)
		assert.Equal(t, "deallocated", status, "VM应该处于释放状态")

		// 最后重新启动VM（恢复初始状态）
		t.Log("恢复VM到运行状态...")
		err = fetcher.VMOperation(ctx, VMOperationStart, vmDetails, nil)
		assert.NoError(t, err)
	})
}
