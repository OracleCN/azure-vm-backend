package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

// VMDetails 包含虚拟机的详细信息
type VMDetails struct {
	// 基本信息
	SubscriptionID string `json:"subscriptionId"`
	ResourceGroup  string `json:"resourceGroup"`
	Name           string `json:"name"`
	ID             string `json:"id"`
	Location       string `json:"location"`
	Size           string `json:"size"`
	State          string `json:"state"`
	Status         string `json:"status"`
	// 网络信息
	PrivateIPs []string `json:"privateIps"`
	PublicIPs  []string `json:"publicIps"`

	// 磁盘信息
	OSDiskSize int32      `json:"osDiskSize"`
	DataDisks  []DiskInfo `json:"dataDisks"`

	// 系统信息
	OSType  string `json:"osType"`  // Windows/Linux
	OSImage string `json:"osImage"` // 完整的镜像信息

	// 标签和其他元数据
	Tags        map[string]string `json:"tags"`
	CreatedTime time.Time         `json:"createdTime"`

	// 配置信息
	NumberOfCores int32  `json:"numberOfCores"`
	MemoryInGB    int32  `json:"memoryInGB"`
	DnsAlias      string `json:"dnsAlias"`
	PublicIPName  string `json:"publicIpName"`
	// 获取时间
	FetchedAt time.Time `json:"fetchedAt"`

	PowerState string `json:"powerState"`
}

// DiskInfo 包含磁盘的详细信息
type DiskInfo struct {
	Name     string `json:"name"`
	SizeGB   int32  `json:"sizeGb"`
	Lun      int32  `json:"lun"`
	DiskType string `json:"diskType"`
}

// VMFetcher 用于获取虚拟机信息的结构体
type VMFetcher struct {
	credentials *Credentials
	logger      *zap.Logger // 改为使用 zap.Logger
	timeout     time.Duration
}

// NewVMFetcher 创建一个新的虚拟机信息获取器
func NewVMFetcher(credentials *Credentials, logger *zap.Logger, timeout time.Duration) *VMFetcher {
	if timeout == 0 {
		timeout = 60 * time.Second // 默认超时时间
	}
	return &VMFetcher{
		credentials: credentials,
		logger:      logger,
		timeout:     timeout,
	}
}

// createAzureCredential 创建Azure凭据对象
func createAzureCredential(creds *Credentials) (*azidentity.ClientSecretCredential, error) {
	if creds == nil {
		return nil, fmt.Errorf("凭据不能为空")
	}

	credential, err := azidentity.NewClientSecretCredential(
		creds.TenantID,
		creds.ClientID,
		creds.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("创建凭据对象失败: %w", err)
	}

	return credential, nil
}

// extractSubscriptionID 从完整的订阅路径中提取订阅ID
func extractSubscriptionID(subscriptionPath string) string {
	// 处理空值
	if subscriptionPath == "" {
		return ""
	}
	// 如果是完整路径，提取订阅ID部分
	if strings.HasPrefix(subscriptionPath, "/subscriptions/") {
		parts := strings.Split(subscriptionPath, "/")
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	// 如果已经是纯订阅ID，直接返回
	return subscriptionPath
}

// FetchVMDetails 获取所有订阅下的虚拟机详细信息
func (f *VMFetcher) FetchVMDetails(ctx context.Context) ([]VMDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	f.logger.Info("开始获取虚拟机详细信息")
	startTime := time.Now()

	// 1. 获取所有可用的订阅
	subscriptionFetcher := NewFetcher(f.credentials, f.logger, f.timeout)
	subscriptions, err := subscriptionFetcher.FetchSubscriptionDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取订阅列表失败: %w", err)
	}

	// 创建凭据
	cred, err := createAzureCredential(f.credentials)
	if err != nil {
		return nil, fmt.Errorf("创建Azure凭据失败: %w", err)
	}

	// 使用通道来管理并发
	vmChan := make(chan VMDetails)
	errChan := make(chan error)
	var wg sync.WaitGroup

	// 收集结果的切片
	var allVMs []VMDetails

	// 启动结果收集goroutine
	done := make(chan bool)
	go func() {
		for vm := range vmChan {
			allVMs = append(allVMs, vm)
		}
		done <- true
	}()

	// 为每个订阅创建一个goroutine
	for _, sub := range subscriptions {
		subscriptionID := extractSubscriptionID(sub.SubscriptionID)
		if subscriptionID == "" {
			f.logger.Error("无效的订阅ID",
				zap.String("rawSubscriptionId", sub.SubscriptionID))
			continue
		}

		wg.Add(1)
		go func(subscription SubscriptionDetail) {
			defer wg.Done()

			vmClient, err := armcompute.NewVirtualMachinesClient(subscription.SubscriptionID, cred, nil)
			if err != nil {
				errChan <- fmt.Errorf("创建虚拟机客户端失败: %w", err)
				return
			}

			pager := vmClient.NewListAllPager(nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					errChan <- fmt.Errorf("获取虚拟机列表失败: %w", err)
					return
				}

				for _, vm := range page.Value {
					vmDetail, err := f.extractVMDetails(ctx, subscription.SubscriptionID, vm, cred)
					if err != nil {
						f.logger.Error("解析虚拟机详情失败",
							zap.String("subscriptionId", subscription.SubscriptionID),
							zap.Error(err))
						continue
					}
					f.logger.Debug("成功解析虚拟机详情",
						zap.String("vmName", vmDetail.Name),
						zap.String("vmId", vmDetail.ID),
						zap.String("state", vmDetail.State))
					vmChan <- vmDetail
				}
			}
		}(sub)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(vmChan)
		close(errChan)
	}()

	// 等待结果收集完成
	<-done

	// 验证结果
	if len(allVMs) == 0 {
		return nil, fmt.Errorf("未找到有效的虚拟机记录")
	}

	f.logger.Info("完成虚拟机详细信息获取",
		zap.Int("totalVMs", len(allVMs)),
		zap.Duration("duration", time.Since(startTime)))

	return allVMs, nil
}

// extractVMDetails 从Azure VM响应中提取详细信息
func (f *VMFetcher) extractVMDetails(ctx context.Context, subscriptionID string, vm *armcompute.VirtualMachine, cred *azidentity.ClientSecretCredential) (VMDetails, error) {
	details := VMDetails{
		SubscriptionID: subscriptionID,
		FetchedAt:      time.Now(),
	}

	// 检查基础对象
	if vm == nil {
		return details, fmt.Errorf("虚拟机对象为空")
	}

	// 处理必需字段
	if vm.ID == nil {
		return details, fmt.Errorf("虚拟机 ID 为空")
	}
	details.ID = *vm.ID
	details.ResourceGroup = extractResourceGroupFromID(details.ID)

	if vm.Name == nil {
		return details, fmt.Errorf("vm 名称为空")
	}
	details.Name = *vm.Name

	if vm.Location == nil {
		return details, fmt.Errorf("vm 位置为空")
	}
	// 获取创建时间
	if vm.Properties != nil && vm.Properties.TimeCreated != nil {
		details.CreatedTime = *vm.Properties.TimeCreated
	}
	details.Location = *vm.Location

	// 处理可选字段
	if vm.Properties != nil {
		// 获取VM大小
		if vm.Properties.HardwareProfile != nil && vm.Properties.HardwareProfile.VMSize != nil {
			details.Size = string(*vm.Properties.HardwareProfile.VMSize)
		}

		// 获取操作系统类型和镜像信息
		if vm.Properties.StorageProfile != nil && vm.Properties.StorageProfile.OSDisk != nil {
			if vm.Properties.StorageProfile.OSDisk.OSType != nil {
				details.OSType = string(*vm.Properties.StorageProfile.OSDisk.OSType)
			}

			// 获取操作系统详细信息
			if vm.Properties.StorageProfile.ImageReference != nil {
				imgRef := vm.Properties.StorageProfile.ImageReference
				var osInfo []string

				if imgRef.Publisher != nil {
					osInfo = append(osInfo, *imgRef.Publisher)
				}
				if imgRef.Offer != nil {
					osInfo = append(osInfo, *imgRef.Offer)
				}
				if imgRef.SKU != nil {
					osInfo = append(osInfo, *imgRef.SKU)
				}
				if imgRef.Version != nil && *imgRef.Version != "latest" {
					osInfo = append(osInfo, *imgRef.Version)
				}

				details.OSImage = strings.Join(osInfo, ":")
			}
		}

		// 获取VM状态
		if vm.Properties.ProvisioningState != nil {
			details.State = *vm.Properties.ProvisioningState
		}

		// 如果有实例视图，获取更详细的状态
		// 首先获取实例视图以获取最新状态
		vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
		if err != nil {
			f.logger.Error("创建VM客户端失败", zap.Error(err))
		} else {
			instanceView, err := vmClient.InstanceView(ctx, details.ResourceGroup, details.Name, nil)
			if err != nil {
				f.logger.Error("获取VM实例视图失败",
					zap.String("vmName", details.Name),
					zap.Error(err))
			} else if instanceView.Statuses != nil {
				// 解析状态信息
				var provisioningState, powerState string
				for _, status := range instanceView.Statuses {
					if status.Code == nil {
						continue
					}

					code := *status.Code
					// 处理配置状态
					if strings.HasPrefix(code, "ProvisioningState/") {
						provisioningState = strings.TrimPrefix(code, "ProvisioningState/")
					}
					// 处理电源状态
					if strings.HasPrefix(code, "PowerState/") {
						powerState = strings.TrimPrefix(code, "PowerState/")
					}
				}

				// 设置状态信息
				details.State = provisioningState // 部署状态
				details.PowerState = powerState   // 电源状态

				// 根据电源状态设置运行状态
				switch powerState {
				case "running":
					details.Status = "Running"
				case "stopped":
					details.Status = "Stopped"
				case "deallocated":
					details.Status = "Deallocated"
				case "starting":
					details.Status = "Starting"
				case "stopping":
					details.Status = "Stopping"
				case "deallocating":
					details.Status = "Deallocating"
				default:
					details.Status = "Unknown"
				}
			}
		}

		// 处理存储配置
		if vm.Properties.StorageProfile != nil {
			// OS磁盘信息
			if vm.Properties.StorageProfile.OSDisk != nil {
				if vm.Properties.StorageProfile.OSDisk.DiskSizeGB != nil {
					details.OSDiskSize = *vm.Properties.StorageProfile.OSDisk.DiskSizeGB
				}
			}

			// 数据磁盘信息
			if vm.Properties.StorageProfile.DataDisks != nil {
				for _, disk := range vm.Properties.StorageProfile.DataDisks {
					diskInfo := DiskInfo{}

					if disk.Name != nil {
						diskInfo.Name = *disk.Name
					}

					if disk.DiskSizeGB != nil {
						diskInfo.SizeGB = *disk.DiskSizeGB
					}

					if disk.Lun != nil {
						diskInfo.Lun = *disk.Lun
					}

					if disk.ManagedDisk != nil && disk.ManagedDisk.StorageAccountType != nil {
						diskInfo.DiskType = string(*disk.ManagedDisk.StorageAccountType)
					}

					// 只添加有效的磁盘信息
					if diskInfo.Name != "" && diskInfo.SizeGB > 0 {
						details.DataDisks = append(details.DataDisks, diskInfo)
					}
				}
			}
		}

		// 处理网络配置
		if vm.Properties.NetworkProfile != nil && vm.Properties.NetworkProfile.NetworkInterfaces != nil {
			// 创建网络客户端
			networkClient, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
			if err != nil {
				f.logger.Error("创建网络客户端失败",
					zap.String("vmName", details.Name),
					zap.Error(err))
			} else {
				for _, nic := range vm.Properties.NetworkProfile.NetworkInterfaces {
					if nic.ID == nil {
						continue
					}

					// 从网络接口ID中提取资源组和名称
					nicResourceGroup := extractResourceGroupFromID(*nic.ID)
					nicName := extractResourceNameFromID(*nic.ID)

					// 获取网络接口详情
					nicResponse, err := networkClient.Get(ctx, nicResourceGroup, nicName, nil)
					if err != nil {
						f.logger.Error("获取网络接口失败",
							zap.String("vmName", details.Name),
							zap.String("nicId", *nic.ID),
							zap.Error(err))
						continue
					}

					// 处理 IP 配置
					if nicResponse.Properties != nil && nicResponse.Properties.IPConfigurations != nil {
						for _, ipConfig := range nicResponse.Properties.IPConfigurations {
							if ipConfig.Properties == nil {
								continue
							}

							// 获取私有 IP
							if ipConfig.Properties.PrivateIPAddress != nil {
								details.PrivateIPs = append(details.PrivateIPs, *ipConfig.Properties.PrivateIPAddress)
								f.logger.Debug("找到子网 IP",
									zap.String("vmName", details.Name),
									zap.String("privateIP", *ipConfig.Properties.PrivateIPAddress))
							}

							// 获取公网 IP（如果有）
							if ipConfig.Properties.PublicIPAddress != nil && ipConfig.Properties.PublicIPAddress.ID != nil {
								pubIPResourceGroup := extractResourceGroupFromID(*ipConfig.Properties.PublicIPAddress.ID)
								pubIPName := extractResourceNameFromID(*ipConfig.Properties.PublicIPAddress.ID)

								details.PublicIPName = pubIPName

								f.logger.Debug("找到公共 IP 资源",
									zap.String("vmName", details.Name),
									zap.String("publicIpName", pubIPName),
									zap.String("resourceGroup", pubIPResourceGroup))
								// 创建公网 IP 客户端
								pubIPClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
								if err != nil {
									f.logger.Error("创建公网IP客户端失败",
										zap.String("vmName", details.Name),
										zap.Error(err))
									continue
								}

								// 获取公网 IP 详情
								pubIP, err := pubIPClient.Get(ctx, pubIPResourceGroup, pubIPName, nil)
								if err != nil {
									f.logger.Error("获取公网IP失败",
										zap.String("vmName", details.Name),
										zap.String("publicIpId", *ipConfig.Properties.PublicIPAddress.ID),
										zap.Error(err))
									continue
								}

								if pubIP.Properties != nil && pubIP.Properties.IPAddress != nil {
									details.PublicIPs = append(details.PublicIPs, *pubIP.Properties.IPAddress)
									f.logger.Debug("找到公共 IP",
										zap.String("vmName", details.Name),
										zap.String("publicIP", *pubIP.Properties.IPAddress))

								}
								if pubIP.Properties.DNSSettings != nil && pubIP.Properties.DNSSettings.Fqdn != nil {
									details.DnsAlias = *pubIP.Properties.DNSSettings.Fqdn
									f.logger.Debug("找到 DNS 别名",
										zap.String("vmName", details.Name),
										zap.String("dnsAlias", details.DnsAlias))
								} else if pubIP.Properties.DNSSettings != nil && pubIP.Properties.DNSSettings.DomainNameLabel != nil {
									// 如果没有完整的 FQDN，但有 domain name label，我们也可以记录它
									domainLabel := *pubIP.Properties.DNSSettings.DomainNameLabel
									if pubIP.Properties.DNSSettings.DomainNameLabel != nil {
										// 构造完整的 FQDN
										location := details.Location
										details.DnsAlias = fmt.Sprintf("%s.%s.cloudapp.azure.com", domainLabel, location)
										f.logger.Debug("根据域名标签构建 DNS 别名",
											zap.String("vmName", details.Name),
											zap.String("dnsAlias", details.DnsAlias))
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 处理标签
	if vm.Tags != nil {
		details.Tags = make(map[string]string)
		for k, v := range vm.Tags {
			if v != nil {
				details.Tags[k] = *v
			}
		}
	}
	// 获取虚拟机大小详情
	if vm.Properties != nil && vm.Properties.HardwareProfile != nil && vm.Properties.HardwareProfile.VMSize != nil {
		sizeClient, err := armcompute.NewVirtualMachineSizesClient(subscriptionID, cred, nil)
		if err != nil {
			f.logger.Error("创建VM规格客户端失败", zap.Error(err))
		} else {
			pager := sizeClient.NewListPager(details.Location, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					f.logger.Error("获取VM规格列表失败", zap.Error(err))
					break
				}

				// 查找匹配的 VM 大小
				vmSize := string(*vm.Properties.HardwareProfile.VMSize)
				for _, size := range page.Value {
					if size.Name != nil && *size.Name == vmSize {
						if size.NumberOfCores != nil {
							details.NumberOfCores = *size.NumberOfCores
						}
						if size.MemoryInMB != nil {
							details.MemoryInGB = *size.MemoryInMB / 1024
						}
						break
					}
				}
			}
		}
	}
	// 记录获取到的VM详情
	f.logger.Debug("提取的虚拟机详细信息",
		zap.String("vmName", details.Name),
		zap.String("location", details.Location),
		zap.String("size", details.Size),
		zap.String("osType", details.OSType),
		zap.String("osImage", details.OSImage),
		zap.String("status", details.Status),
		zap.String("powerState", details.PowerState),
		zap.String("provisioningState", details.State),
		zap.Int32("osDiskSize", details.OSDiskSize),
		zap.Int("dataDisksCount", len(details.DataDisks)))
	// json 格式化输出details 对象
	return details, nil
}

// SetVMDNSLabel 为虚拟机设置 DNS 名称标签，返回设置后的FQDN
func (f *VMFetcher) SetVMDNSLabel(ctx context.Context, subscriptionID, resourceGroup, publicIPName, dnsLabel string) (string, error) {
	// 创建凭据
	cred, err := createAzureCredential(f.credentials)
	if err != nil {
		return "", fmt.Errorf("创建Azure凭据失败: %w", err)
	}
	// 创建公共IP客户端
	pipClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return "", fmt.Errorf("创建公共IP客户端失败: %w", err)
	}
	// 获取当前公共IP配置
	pipResponse, err := pipClient.Get(ctx, resourceGroup, publicIPName, nil)
	if err != nil {
		return "", fmt.Errorf("获取公共IP配置失败: %w", err)
	}
	// 更新DNS设置
	publicIP := pipResponse.PublicIPAddress
	location := *publicIP.Location // 保存位置信息用于构造FQDN
	if publicIP.Properties == nil {
		publicIP.Properties = &armnetwork.PublicIPAddressPropertiesFormat{}
	}
	publicIP.Properties.DNSSettings = &armnetwork.PublicIPAddressDNSSettings{
		DomainNameLabel: &dnsLabel,
	}
	// 应用更新
	poller, err := pipClient.BeginCreateOrUpdate(ctx, resourceGroup, publicIPName, publicIP, nil)
	if err != nil {
		return "", fmt.Errorf("开始更新DNS标签失败: %w", err)
	}
	_, err = poller.Poll(ctx)
	if err != nil {
		return "", fmt.Errorf("更新DNS标签失败: %w", err)
	}
	// 验证更新
	updatedPIP, err := pipClient.Get(ctx, resourceGroup, publicIPName, nil)
	if err != nil {
		return "", fmt.Errorf("验证DNS标签更新失败: %w", err)
	}
	// 检查更新是否成功
	if updatedPIP.PublicIPAddress.Properties == nil ||
		updatedPIP.PublicIPAddress.Properties.DNSSettings == nil ||
		updatedPIP.PublicIPAddress.Properties.DNSSettings.DomainNameLabel == nil ||
		*updatedPIP.PublicIPAddress.Properties.DNSSettings.DomainNameLabel != dnsLabel {
		return "", fmt.Errorf("DNS标签未成功更新")
	}
	// 返回完整的FQDN
	fqdn := fmt.Sprintf("%s.%s.cloudapp.azure.com", dnsLabel, location)
	f.logger.Info("DNS名称设置成功",
		zap.String("publicIPName", publicIPName),
		zap.String("dnsLabel", dnsLabel),
		zap.String("fqdn", fqdn))

	return fqdn, nil
}

// extractResourceGroupFromID 从资源ID中提取资源组名称
func extractResourceGroupFromID(id string) string {
	// ID格式: /subscriptions/{subID}/resourceGroups/{resourceGroup}/...
	parts := strings.Split(id, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
func extractResourceNameFromID(id string) string {
	if id == "" {
		return ""
	}
	parts := strings.Split(id, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
