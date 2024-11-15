package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testLogger *zap.Logger
	testCred   *AzureCredential
	tempDir    string
)

func init() {
	// 初始化日志
	logger, _ := zap.NewDevelopment()
	testLogger = logger

	// 从环境变量获取测试凭据
	testCred = &AzureCredential{
		TenantID:     "d6fcb345-610d-4ed2-ac0a-cfa940f9fa5f",
		ClientID:     "d91bca22-19fe-4c8a-82f1-c0bcd2cdbb94",
		ClientSecret: "9t.8Q~I3nIFtGxr4elqlEPB2TXdfDarrwSRbpaXk",
	}

	// 创建临时目录
	tempDir = filepath.Join(".", "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		panic(fmt.Sprintf("创建临时目录失败: %v", err))
	}
}

// writeJSON 将数据写入JSON文件
func writeJSON(filename string, data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	filePath := filepath.Join(tempDir, filename)
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func TestRegionFetcher_GetRegions(t *testing.T) {
	if testCred.TenantID == "" || testCred.ClientID == "" || testCred.ClientSecret == "" {
		t.Skip("未设置Azure凭据环境变量，跳过测试")
	}

	ctx := context.Background()
	fetcher := NewRegionFetcher(testLogger, 3, 30*time.Second)

	subscriptionID := "1802f3c5-7ca4-4109-b437-3e34932602c9"
	if subscriptionID == "" {
		t.Fatal("未设置AZURE_SUBSCRIPTION_ID环境变量")
	}

	// 测试获取区域列表
	regions, err := fetcher.GetRegions(ctx, testCred, subscriptionID)
	assert.NoError(t, err)
	assert.NotEmpty(t, regions)

	// 写入完整的区域信息
	err = writeJSON("regions_full.json", regions)
	if err != nil {
		t.Errorf("写入区域信息失败: %v", err)
	} else {
		t.Logf("完整的区域信息已写入: %s", filepath.Join(tempDir, "regions_full.json"))
	}

	// 写入统计信息
	regionStats := struct {
		TotalCount int      `json:"total_count"`
		Regions    []string `json:"region_names"`
	}{
		TotalCount: len(regions),
		Regions:    make([]string, 0, len(regions)),
	}

	for _, region := range regions {
		regionStats.Regions = append(regionStats.Regions, region.Name)
	}

	err = writeJSON("regions_stats.json", regionStats)
	if err != nil {
		t.Errorf("写入区域统计信息失败: %v", err)
	} else {
		t.Logf("区域统计信息已写入: %s", filepath.Join(tempDir, "regions_stats.json"))
	}
}

func TestRegionFetcher_IsRegionAvailable(t *testing.T) {
	if testCred.TenantID == "" || testCred.ClientID == "" || testCred.ClientSecret == "" {
		t.Skip("未设置Azure凭据环境变量，跳过测试")
	}

	ctx := context.Background()
	fetcher := NewRegionFetcher(testLogger, 3, 30*time.Second)
	subscriptionID := "1802f3c5-7ca4-4109-b437-3e34932602c9"

	tests := []struct {
		name       string
		regionName string
		wantErr    bool
	}{
		{
			name:       "测试有效区域",
			regionName: "eastus",
			wantErr:    false,
		},
		{
			name:       "测试无效区域",
			regionName: "invalid-region",
			wantErr:    false,
		},
	}

	var results []struct {
		TestName   string `json:"test_name"`
		RegionName string `json:"region_name"`
		Available  bool   `json:"available"`
		Error      string `json:"error,omitempty"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available, err := fetcher.IsRegionAvailable(ctx, testCred, subscriptionID, tt.regionName)

			result := struct {
				TestName   string `json:"test_name"`
				RegionName string `json:"region_name"`
				Available  bool   `json:"available"`
				Error      string `json:"error,omitempty"`
			}{
				TestName:   tt.name,
				RegionName: tt.regionName,
				Available:  available,
			}

			if err != nil {
				result.Error = err.Error()
			}

			results = append(results, result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// 写入可用性测试结果
	err := writeJSON("region_availability_tests.json", results)
	if err != nil {
		t.Errorf("写入区域可用性测试结果失败: %v", err)
	} else {
		t.Logf("区域可用性测试结果已写入: %s", filepath.Join(tempDir, "region_availability_tests.json"))
	}
}

func TestRegionFetcher_GetRegions_WithTimeout(t *testing.T) {
	if testCred.TenantID == "" || testCred.ClientID == "" || testCred.ClientSecret == "" {
		t.Skip("未设置Azure凭据环境变量，跳过测试")
	}

	ctx := context.Background()
	fetcher := NewRegionFetcher(testLogger, 1, 1*time.Millisecond)
	subscriptionID := "1802f3c5-7ca4-4109-b437-3e34932602c9"

	_, err := fetcher.GetRegions(ctx, testCred, subscriptionID)

	result := struct {
		TestName string `json:"test_name"`
		Error    string `json:"error,omitempty"`
	}{
		TestName: "超时测试",
		Error:    err.Error(),
	}

	err = writeJSON("timeout_test.json", result)
	if err != nil {
		t.Errorf("写入超时测试结果失败: %v", err)
	} else {
		t.Logf("超时测试结果已写入: %s", filepath.Join(tempDir, "timeout_test.json"))
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "超时")
}
