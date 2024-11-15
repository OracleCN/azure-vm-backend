package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"go.uber.org/zap"
)

// RegionFetcher 区域信息获取器
type RegionFetcher struct {
	logger  *zap.Logger
	retries int           // 重试次数
	timeout time.Duration // 超时时间
}

// NewRegionFetcher 创建区域信息获取器
func NewRegionFetcher(logger *zap.Logger, retries int, timeout time.Duration) *RegionFetcher {
	if retries <= 0 {
		retries = 3 // 默认重试3次
	}
	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}
	return &RegionFetcher{
		logger:  logger,
		retries: retries,
		timeout: timeout,
	}
}

// RegionInfo 区域信息
type RegionInfo struct {
	Name        string            // 区域名称
	DisplayName string            // 显示名称
	Location    string            // 地理位置
	Status      string            // 状态
	Metadata    map[string]string // 元数据
}

// GetRegions 获取指定订阅下所有可用区域
func (f *RegionFetcher) GetRegions(ctx context.Context, cred *AzureCredential, subscriptionID string) ([]RegionInfo, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	var regions []RegionInfo
	var lastErr error

	// 重试逻辑
	for attempt := 0; attempt < f.retries; attempt++ {
		if attempt > 0 {
			f.logger.Warn("重试获取区域列表",
				zap.String("subscription_id", subscriptionID),
				zap.Int("attempt", attempt+1),
				zap.Error(lastErr),
			)
			// 重试等待
			time.Sleep(time.Second * time.Duration(attempt+1))
		}

		regions, lastErr = f.fetchRegions(ctx, cred, subscriptionID)
		if lastErr == nil {
			return regions, nil
		}

		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return nil, fmt.Errorf("获取区域列表超时: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("获取区域列表失败(已重试%d次): %w", f.retries, lastErr)
}

// fetchRegions 实际获取区域列表的核心逻辑
func (f *RegionFetcher) fetchRegions(ctx context.Context, cred *AzureCredential, subscriptionID string) ([]RegionInfo, error) {
	credential, err := cred.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armsubscriptions.NewClient(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建订阅客户端失败: %w", err)
	}

	pager := client.NewListLocationsPager(subscriptionID, nil)
	var regions []RegionInfo

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取区域列表失败: %w", err)
		}

		for _, location := range page.Value {
			if location == nil || location.Name == nil {
				f.logger.Warn("跳过无效的区域信息")
				continue
			}

			region := RegionInfo{
				Name:        *location.Name,
				DisplayName: *location.DisplayName,
				Location:    *location.RegionalDisplayName,
			}

			if location.Metadata != nil {
				region.Metadata = make(map[string]string)
				if regionType := location.Metadata.RegionType; regionType != nil {
					region.Status = string(*regionType)
					region.Metadata["regionType"] = string(*regionType)
				}
				if category := location.Metadata.RegionCategory; category != nil {
					region.Metadata["regionCategory"] = string(*category)
				}
				if geography := location.Metadata.Geography; geography != nil {
					region.Metadata["geography"] = string(*geography)
				}
				if group := location.Metadata.GeographyGroup; group != nil {
					region.Metadata["geographyGroup"] = string(*group)
				}
			}

			regions = append(regions, region)
		}
	}

	f.logger.Info("成功获取区域列表",
		zap.String("subscription_id", subscriptionID),
		zap.Int("region_count", len(regions)),
	)

	return regions, nil
}

// IsRegionAvailable 检查指定区域是否可用
func (f *RegionFetcher) IsRegionAvailable(ctx context.Context, cred *AzureCredential, subscriptionID string, regionName string) (bool, error) {
	regions, err := f.GetRegions(ctx, cred, subscriptionID)
	if err != nil {
		return false, fmt.Errorf("检查区域可用性失败: %w", err)
	}

	for _, region := range regions {
		if region.Name == regionName {
			// 检查区域状态
			if region.Status == "Physical" || region.Status == "Logical" {
				return true, nil
			}
			return false, nil
		}
	}

	return false, nil
}
