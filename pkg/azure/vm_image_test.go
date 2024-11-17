package azure

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestListSupportedOSImages(t *testing.T) {
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

	fetcher := NewVMImageFetcher(
		"1802f3c5-7ca4-4109-b437-3e34932602c9",
		cred,
		logger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	location := "eastasia"

	t.Run("列举支持的操作系统镜像", func(t *testing.T) {
		// 定义所有支持的操作系统镜像
		osImages := struct {
			linux   []struct{ publisher, offer, sku, desc string }
			windows []struct{ publisher, offer, sku, desc string }
		}{
			linux: []struct{ publisher, offer, sku, desc string }{
				// Ubuntu 系列
				{"Canonical", "0001-com-ubuntu-server-jammy", "22_04-lts-gen2", "Ubuntu 22.04 LTS"},
				{"Canonical", "0001-com-ubuntu-server-focal", "20_04-lts-gen2", "Ubuntu 20.04 LTS"},
				{"Canonical", "0001-com-ubuntu-minimal-jammy", "minimal-22_04-lts-gen2", "Ubuntu 22.04 LTS Minimal"},

				// RHEL 系列
				{"RedHat", "RHEL", "8_8-gen2", "Red Hat Enterprise Linux 8.8"},
				{"RedHat", "RHEL", "9_2-gen2", "Red Hat Enterprise Linux 9.2"},

				// CentOS 系列
				{"OpenLogic", "CentOS", "7_9-gen2", "CentOS 7.9"},
				{"OpenLogic", "CentOS", "8_5-gen2", "CentOS 8.5"},

				// Debian 系列
				{"debian", "debian-11", "11-gen2", "Debian 11"},
				{"debian", "debian-12", "12-gen2", "Debian 12"},

				// SUSE 系列
				{"SUSE", "sles-15-sp5", "gen2", "SUSE Linux Enterprise 15 SP5"},
				{"SUSE", "opensuse-leap-15-5", "gen2", "openSUSE Leap 15.5"},

				// Oracle Linux
				{"Oracle", "Oracle-Linux", "ol85-gen2", "Oracle Linux 8.5"},
				{"Oracle", "Oracle-Linux", "ol84-gen2", "Oracle Linux 8.4"},

				// Alma Linux
				{"almalinux", "almalinux", "8-gen2", "Alma Linux 8"},
				{"almalinux", "almalinux", "9-gen2", "Alma Linux 9"},

				// Rocky Linux
				{"rocky-linux", "rocky-linux", "8-gen2", "Rocky Linux 8"},
				{"rocky-linux", "rocky-linux", "9-gen2", "Rocky Linux 9"},

				// 其他专业发行版
				{"kinvolk", "flatcar-container-linux-free", "stable-gen2", "Flatcar Linux (容器优化)"},
				{"microsoftcblmariner", "cbl-mariner", "2-gen2", "Microsoft CBL-Mariner 2.0"},
				{"vmware-inc", "photon-os", "4_0-gen2", "VMware Photon OS 4.0"},
				{"tunnelshield", "fedora", "fedora-38-gen2", "Fedora 38"},
				{"kali-linux", "kali", "kali-2023-4", "Kali Linux 2023.4"},
			},
			windows: []struct{ publisher, offer, sku, desc string }{
				// Windows Server
				{"MicrosoftWindowsServer", "WindowsServer", "2022-datacenter-g2", "Windows Server 2022 Datacenter"},
				{"MicrosoftWindowsServer", "WindowsServer", "2019-datacenter-gensecond", "Windows Server 2019 Datacenter"},
				{"MicrosoftWindowsServer", "WindowsServer", "2016-datacenter-gensecond", "Windows Server 2016 Datacenter"},

				// Windows 11/10
				{"MicrosoftWindowsDesktop", "Windows-11", "win11-22h2-pro", "Windows 11 Pro"},
				{"MicrosoftWindowsDesktop", "Windows-10", "win10-22h2-pro", "Windows 10 Pro"},

				// Windows Server 容器版本
				{"MicrosoftWindowsServer", "WindowsServer", "2022-datacenter-core-g2", "Windows Server 2022 Core"},
				{"MicrosoftWindowsServer", "WindowsServer", "2019-datacenter-core-gensecond", "Windows Server 2019 Core"},
			},
		}

		// 测试 Linux 发行版
		t.Log("\n=== 支持的 Linux 发行版 ===")
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 5)
		var mu sync.Mutex

		for _, img := range osImages.linux {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(img struct{ publisher, offer, sku, desc string }) {
				defer wg.Done()
				defer func() { <-semaphore }()

				versions, err := fetcher.ListVersions(ctx, location, img.publisher, img.offer, img.sku)
				if err != nil || len(versions) == 0 {
					mu.Lock()
					t.Logf("❌ %s (不可用)", img.desc)
					mu.Unlock()
					return
				}

				latestVersion := versions[len(versions)-1]
				mu.Lock()
				t.Logf("✅ %s", img.desc)
				t.Logf("   发布者: %s", img.publisher)
				t.Logf("   产品: %s", img.offer)
				t.Logf("   SKU: %s", img.sku)
				t.Logf("   最新版本: %s", latestVersion)
				mu.Unlock()
			}(img)
		}
		wg.Wait()

		// 测试 Windows 发行版
		t.Log("\n=== 支持的 Windows 发行版 ===")
		for _, img := range osImages.windows {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(img struct{ publisher, offer, sku, desc string }) {
				defer wg.Done()
				defer func() { <-semaphore }()

				versions, err := fetcher.ListVersions(ctx, location, img.publisher, img.offer, img.sku)
				if err != nil || len(versions) == 0 {
					mu.Lock()
					t.Logf("❌ %s (不可用)", img.desc)
					mu.Unlock()
					return
				}

				latestVersion := versions[len(versions)-1]
				mu.Lock()
				t.Logf("✅ %s", img.desc)
				t.Logf("   发布者: %s", img.publisher)
				t.Logf("   产品: %s", img.offer)
				t.Logf("   SKU: %s", img.sku)
				t.Logf("   最新版本: %s", latestVersion)
				mu.Unlock()
			}(img)
		}
		wg.Wait()
	})
}
