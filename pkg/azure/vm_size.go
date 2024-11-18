package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"go.uber.org/zap"
)

// VMSizeFetcher 用于获取虚拟机规格信息
type VMSizeFetcher struct {
	subscriptionID string
	credentials    *AzureCredential
	logger         *zap.Logger
}

// NewVMSizeFetcher 创建VMSizeFetcher实例
func NewVMSizeFetcher(subscriptionID string, credentials *AzureCredential, logger *zap.Logger) *VMSizeFetcher {
	return &VMSizeFetcher{
		subscriptionID: subscriptionID,
		credentials:    credentials,
		logger:         logger,
	}
}

// VMSizeInfo 包含虚拟机规格的详细信息
type VMSizeInfo struct {
	Name         string
	Location     string
	Cores        int
	MemoryGB     float64
	MaxDataDisks int
	OSDiskSizeGB int
	Category     string
	Family       string
	PricePerHour float64
	Currency     string
}

// ListSizes 获取指定位置的虚拟机规格列表
func (f *VMSizeFetcher) ListSizes(ctx context.Context, location string) ([]*VMSizeInfo, error) {
	credential, err := f.credentials.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("获取认证对象失败: %w", err)
	}

	client, err := armcompute.NewVirtualMachineSizesClient(f.subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("创建规格客户端失败: %w", err)
	}

	pager := client.NewListPager(location, nil)
	var sizes []*VMSizeInfo

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取规格列表失败: %w", err)
		}

		for _, size := range page.Value {
			if size.Name == nil {
				continue
			}

			sizeInfo := &VMSizeInfo{
				Name:     *size.Name,
				Location: location,
			}

			// 填充其他字段
			if size.NumberOfCores != nil {
				sizeInfo.Cores = int(*size.NumberOfCores)
			}
			if size.MemoryInMB != nil {
				sizeInfo.MemoryGB = float64(*size.MemoryInMB) / 1024
			}
			if size.MaxDataDiskCount != nil {
				sizeInfo.MaxDataDisks = int(*size.MaxDataDiskCount)
			}
			if size.OSDiskSizeInMB != nil {
				sizeInfo.OSDiskSizeGB = int(*size.OSDiskSizeInMB) / 1024
			}

			// 解析规格族和类别
			sizeInfo.Family = extractVMFamily(*size.Name)
			sizeInfo.Category = categorizeVMSize(*size.Name)

			sizes = append(sizes, sizeInfo)
		}
	}

	return sizes, nil
}

// extractVMFamily 提取VM规格族
// 例如: Standard_D2s_v3 -> Dsv3
func extractVMFamily(sizeName string) string {
	// 常见的规格族映射
	familyPatterns := map[string]string{
		"Standard_B":  "B",    // Burstable
		"Standard_D":  "D",    // General Purpose
		"Standard_Ds": "Dsv3", // General Purpose with Premium Storage
		"Standard_E":  "E",    // Memory Optimized
		"Standard_Es": "Esv3", // Memory Optimized with Premium Storage
		"Standard_F":  "F",    // Compute Optimized
		"Standard_Fs": "Fsv2", // Compute Optimized with Premium Storage
		"Standard_H":  "H",    // High Performance Compute
		"Standard_L":  "L",    // Storage Optimized
		"Standard_M":  "M",    // Memory Optimized
		"Standard_N":  "N",    // GPU
		"Standard_NC": "NC",   // GPU Compute
		"Standard_ND": "ND",   // GPU
		"Standard_NV": "NV",   // GPU Visualization
	}

	// 遍历规格族映射
	for prefix, family := range familyPatterns {
		if strings.HasPrefix(sizeName, prefix) {
			// 处理版本号
			if strings.Contains(sizeName, "_v") {
				version := strings.Split(sizeName, "_v")[1]
				return family + "v" + version
			}
			return family
		}
	}

	// 如果没有匹配到，返回 Unknown
	return "Unknown"
}

// categorizeVMSize 对VM规格进行分类
func categorizeVMSize(sizeName string) string {
	// 基于规格名称的分类规则
	switch {
	case strings.HasPrefix(sizeName, "Standard_B"):
		return "Burstable"
	case strings.HasPrefix(sizeName, "Standard_D") || strings.HasPrefix(sizeName, "Standard_Ds"):
		return "General Purpose"
	case strings.HasPrefix(sizeName, "Standard_E") || strings.HasPrefix(sizeName, "Standard_Es"):
		return "Memory Optimized"
	case strings.HasPrefix(sizeName, "Standard_F") || strings.HasPrefix(sizeName, "Standard_Fs"):
		return "Compute Optimized"
	case strings.HasPrefix(sizeName, "Standard_H"):
		return "High Performance Computing"
	case strings.HasPrefix(sizeName, "Standard_L"):
		return "Storage Optimized"
	case strings.HasPrefix(sizeName, "Standard_M"):
		return "Memory Optimized"
	case strings.HasPrefix(sizeName, "Standard_N"):
		return "GPU"
	case strings.HasPrefix(sizeName, "Standard_NC"):
		return "GPU Compute"
	case strings.HasPrefix(sizeName, "Standard_ND"):
		return "GPU"
	case strings.HasPrefix(sizeName, "Standard_NV"):
		return "GPU Visualization"
	default:
		return "General Purpose"
	}
}
