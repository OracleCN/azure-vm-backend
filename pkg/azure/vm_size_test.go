package azure

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestVMSizeFetcher_ListSizes(t *testing.T) {
	// 跳过集成测试，除非显式启用
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 初始化测试所需的凭据
	cred := &AzureCredential{
		TenantID:     "d6fcb345-610d-4ed2-ac0a-cfa940f9fa5f",
		ClientID:     "d91bca22-19fe-4c8a-82f1-c0bcd2cdbb94",
		ClientSecret: "9t.8Q~I3nIFtGxr4elqlEPB2TXdfDarrwSRbpaXk",
	}

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	fetcher := NewVMSizeFetcher(
		"1802f3c5-7ca4-4109-b437-3e34932602c9",
		cred,
		logger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("获取 eastasia 区域的VM规格列表", func(t *testing.T) {
		sizes, err := fetcher.ListSizes(ctx, "eastasia")
		if err != nil {
			t.Fatalf("获取VM规格列表失败: %v", err)
		}

		if len(sizes) == 0 {
			t.Fatal("未获取到任何VM规格")
		}

		commonSizes := map[string]bool{
			"Standard_B2s":    false,
			"Standard_D2s_v3": false,
			"Standard_D4s_v3": false,
			"Standard_E2s_v3": false,
			"Standard_F2s_v2": false,
		}

		t.Log("\n=== VM规格信息示例 ===")
		for _, size := range sizes {
			if _, isCommon := commonSizes[size.Name]; isCommon {
				commonSizes[size.Name] = true
				t.Logf("\n规格名称: %s", size.Name)
				t.Logf("CPU核心数: %d", size.Cores)
				t.Logf("内存(GB): %.1f", size.MemoryGB)
				t.Logf("最大数据磁盘数: %d", size.MaxDataDisks)
				t.Logf("系统盘大小(GB): %d", size.OSDiskSizeGB)
				t.Logf("规格族: %s", size.Family)
				t.Logf("类别: %s", size.Category)
			}
		}

		// 验证是否找到了所有常见规格
		for sizeName, found := range commonSizes {
			if !found {
				t.Logf("警告: 未找到常见规格 %s", sizeName)
			}
		}

		// 打印统计信息
		var (
			totalCores    int
			totalMemoryGB float64
			families      = make(map[string]int)
			categories    = make(map[string]int)
		)

		for _, size := range sizes {
			totalCores += size.Cores
			totalMemoryGB += size.MemoryGB
			families[size.Family]++
			categories[size.Category]++
		}

		t.Logf("\n=== 统计信息 ===")
		t.Logf("总规格数量: %d", len(sizes))
		t.Logf("平均核心数: %.1f", float64(totalCores)/float64(len(sizes)))
		t.Logf("平均内存(GB): %.1f", totalMemoryGB/float64(len(sizes)))

		t.Log("\n规格族分布:")
		for family, count := range families {
			t.Logf("%s: %d个规格", family, count)
		}

		t.Log("\n类别分布:")
		for category, count := range categories {
			t.Logf("%s: %d个规格", category, count)
		}
	})
}
