package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"go.uber.org/zap"
)

// VMImageFetcher 用于获取虚拟机镜像信息
type VMImageFetcher struct {
	subscriptionID string
	credentials    *AzureCredential
	logger         *zap.Logger
}

// NewVMImageFetcher 创建VMImageFetcher实例
func NewVMImageFetcher(subscriptionID string, credentials *AzureCredential, logger *zap.Logger) *VMImageFetcher {
	return &VMImageFetcher{
		subscriptionID: subscriptionID,
		credentials:    credentials,
		logger:         logger,
	}
}

// VMImageInfo 包含虚拟机镜像的详细信息
type VMImageInfo struct {
	Publisher   string
	Offer       string
	SKU         string
	Version     string
	OSType      string
	DisplayName string
	Description string
}

// ListPublishers 获取指定位置的发布者列表
func (f *VMImageFetcher) ListPublishers(ctx context.Context, location string) ([]string, error) {
	credential, err := f.credentials.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armcompute.NewVirtualMachineImagesClient(f.subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建镜像客户端失败: %w", err)
	}

	result, err := client.ListPublishers(ctx, location, nil)
	if err != nil {
		return nil, fmt.Errorf("获取发布者列表失败: %w", err)
	}

	var publishers []string
	for _, pub := range result.VirtualMachineImageResourceArray {
		if pub.Name != nil {
			publishers = append(publishers, *pub.Name)
		}
	}

	return publishers, nil
}

// ListOffers 获取指定发布者的产品列表
func (f *VMImageFetcher) ListOffers(ctx context.Context, location, publisher string) ([]string, error) {
	credential, err := f.credentials.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armcompute.NewVirtualMachineImagesClient(f.subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建镜像客户端失败: %w", err)
	}

	result, err := client.ListOffers(ctx, location, publisher, nil)
	if err != nil {
		return nil, fmt.Errorf("获取产品列表失败: %w", err)
	}

	var offers []string
	for _, offer := range result.VirtualMachineImageResourceArray {
		if offer.Name != nil {
			offers = append(offers, *offer.Name)
		}
	}

	return offers, nil
}

// ListSKUs 获取指定产品的SKU列表
func (f *VMImageFetcher) ListSKUs(ctx context.Context, location, publisher, offer string) ([]string, error) {
	credential, err := f.credentials.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armcompute.NewVirtualMachineImagesClient(f.subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建镜像客户端失败: %w", err)
	}

	result, err := client.ListSKUs(ctx, location, publisher, offer, nil)
	if err != nil {
		return nil, fmt.Errorf("获取SKU列表失败: %w", err)
	}

	var skus []string
	for _, sku := range result.VirtualMachineImageResourceArray {
		if sku.Name != nil {
			skus = append(skus, *sku.Name)
		}
	}

	return skus, nil
}

// ListVersions 获取指定SKU的版本列表
func (f *VMImageFetcher) ListVersions(ctx context.Context, location, publisher, offer, sku string) ([]string, error) {
	credential, err := f.credentials.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armcompute.NewVirtualMachineImagesClient(f.subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建镜像客户端失败: %w", err)
	}

	result, err := client.List(ctx, location, publisher, offer, sku, nil)
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}

	var versions []string
	for _, version := range result.VirtualMachineImageResourceArray {
		if version.Name != nil {
			versions = append(versions, *version.Name)
		}
	}

	return versions, nil
}

// GetImage 获取指定镜像的详细信息
func (f *VMImageFetcher) GetImage(ctx context.Context, location, publisher, offer, sku, version string) (*VMImageInfo, error) {
	f.logger.Debug("开始获取镜像详情",
		zap.String("location", location),
		zap.String("publisher", publisher),
		zap.String("offer", offer),
		zap.String("sku", sku),
		zap.String("version", version))

	// 获取认证对象
	credential, err := f.credentials.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armcompute.NewVirtualMachineImagesClient(f.subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建镜像客户端失败: %w", err)
	}

	image, err := client.Get(ctx, location, publisher, offer, sku, version, nil)
	if err != nil {
		return nil, fmt.Errorf("获取镜像详情失败: %w", err)
	}

	imageInfo := &VMImageInfo{
		Publisher: publisher,
		Offer:     offer,
		SKU:       sku,
		Version:   version,
	}

	if image.Properties != nil {
		if image.Properties.OSDiskImage != nil && image.Properties.OSDiskImage.OperatingSystem != nil {
			imageInfo.OSType = string(*image.Properties.OSDiskImage.OperatingSystem)
		}
	}

	f.logger.Debug("成功获取镜像详情",
		zap.String("publisher", publisher),
		zap.String("offer", offer),
		zap.String("sku", sku),
		zap.String("osType", imageInfo.OSType))

	return imageInfo, nil
}

// 添加热门镜像配置结构
type ImageSpec struct {
	Publisher string
	Offer     string
	SKU       string
	Desc      string
}

// 定义热门镜像列表
var popularImages = []ImageSpec{
	// Ubuntu 系列
	{"Canonical", "0001-com-ubuntu-server-jammy", "22_04-lts-gen2", "Ubuntu 22.04 LTS"},
	{"Canonical", "0001-com-ubuntu-server-focal", "20_04-lts-gen2", "Ubuntu 20.04 LTS"},
	{"Canonical", "0001-com-ubuntu-minimal-jammy", "minimal-22_04-lts-gen2", "Ubuntu 22.04 LTS Minimal"},

	// CentOS 系列 (全部可用)
	{"OpenLogic", "CentOS", "7_9-gen2", "CentOS 7.9"},
	{"OpenLogic", "CentOS", "8_5-gen2", "CentOS 8.5"},

	// Debian 系列 (全部可用)
	{"debian", "debian-11", "11-gen2", "Debian 11"},
	{"debian", "debian-12", "12-gen2", "Debian 12"},

	// SUSE 系列 (全部可用)
	{"SUSE", "sles-15-sp5", "gen2", "SUSE Linux Enterprise 15 SP5"},
	{"SUSE", "opensuse-leap-15-5", "gen2", "openSUSE Leap 15.5"},

	// Alma Linux (全部可用)
	{"almalinux", "almalinux", "8-gen2", "Alma Linux 8"},
	{"almalinux", "almalinux", "9-gen2", "Alma Linux 9"},

	// 容器和特殊用途 Linux (部分可用)
	{"kinvolk", "flatcar-container-linux-free", "stable-gen2", "Flatcar Linux (容器优化)"},
	{"MicrosoftWindowsServer", "WindowsServer", "2022-datacenter-g2", "Windows Server 2022 Datacenter"},
	{"MicrosoftWindowsServer", "WindowsServer", "2019-datacenter-gensecond", "Windows Server 2019 Datacenter"},
	{"MicrosoftWindowsServer", "WindowsServer", "2016-datacenter-gensecond", "Windows Server 2016 Datacenter"},

	// Windows Desktop (全部可用)
	{"MicrosoftWindowsDesktop", "Windows-11", "win11-22h2-pro", "Windows 11 Pro"},
	{"MicrosoftWindowsDesktop", "Windows-10", "win10-22h2-pro", "Windows 10 Pro"},

	// Windows Server Core
	{"MicrosoftWindowsServer", "WindowsServer", "2022-datacenter-core-g2", "Windows Server 2022 Core"},
}

// SyncImages 同步指定位置的热门镜像信息
func (f *VMImageFetcher) SyncImages(ctx context.Context, location string) ([]*VMImageInfo, error) {
	f.logger.Info("开始同步热门镜像信息", zap.String("location", location))

	// 创建一个带缓冲的通道，用于收集结果
	resultChan := make(chan *VMImageInfo, len(popularImages))
	errorChan := make(chan error, len(popularImages))

	// 使用 WaitGroup 等待所有协程完成
	var wg sync.WaitGroup

	// 限制并发数量
	semaphore := make(chan struct{}, 5) // 最多5个并发

	// 启动协程处理每个镜像
	for _, spec := range popularImages {
		wg.Add(1)
		go func(spec ImageSpec) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
			}

			// 获取版本列表
			versions, err := f.ListVersions(ctx, location, spec.Publisher, spec.Offer, spec.SKU)
			if err != nil {
				f.logger.Error("获取版本列表失败",
					zap.String("publisher", spec.Publisher),
					zap.String("offer", spec.Offer),
					zap.String("sku", spec.SKU),
					zap.Error(err))
				errorChan <- err
				return
			}

			if len(versions) > 0 {
				latestVersion := versions[len(versions)-1]
				image, err := f.GetImage(ctx, location, spec.Publisher, spec.Offer, spec.SKU, latestVersion)
				if err != nil {
					f.logger.Error("获取镜像详情失败",
						zap.String("publisher", spec.Publisher),
						zap.String("offer", spec.Offer),
						zap.String("sku", spec.SKU),
						zap.String("version", latestVersion),
						zap.Error(err))
					errorChan <- err
					return
				}

				// 添加描述信息
				image.DisplayName = spec.Desc
				image.Description = spec.Desc

				f.logger.Info("成功获取镜像信息",
					zap.String("publisher", spec.Publisher),
					zap.String("offer", spec.Offer),
					zap.String("sku", spec.SKU),
					zap.String("version", latestVersion),
					zap.String("description", spec.Desc))

				resultChan <- image
			}
		}(spec)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 收集结果
	var allImages []*VMImageInfo
	var errors []error

	// 从通道中获取结果和错误
	for {
		select {
		case img, ok := <-resultChan:
			if !ok {
				// 通道已关闭，所有结果都已收集
				if len(errors) > 0 {
					f.logger.Warn("同步过程中发生错误",
						zap.Int("error_count", len(errors)),
						zap.Error(errors[0]))
				}
				return allImages, nil
			}
			if img != nil {
				allImages = append(allImages, img)
			}
		case err, ok := <-errorChan:
			if !ok {
				continue
			}
			if err != nil {
				errors = append(errors, err)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
