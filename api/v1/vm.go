package v1

import "time"

// VMCreateParams 创建虚拟机的参数
type VMCreateParams struct {
	SubscriptionID string            `json:"subscriptionId"`
	ResourceGroup  string            `json:"resourceGroup"`
	Name           string            `json:"name"`
	Location       string            `json:"location"`
	Size           string            `json:"size"`
	OSType         string            `json:"osType"`
	OSDiskSize     int               `json:"osDiskSize"`
	DataDisks      []VMDiskParams    `json:"dataDisks"`
	NetworkConfig  VMNetworkParams   `json:"networkConfig"`
	Tags           map[string]string `json:"tags"`
}

// VMDiskParams 虚拟机磁盘参数
type VMDiskParams struct {
	Name     string `json:"name"`
	SizeGB   int    `json:"sizeGb"`
	DiskType string `json:"diskType"`
}

// VMNetworkParams 虚拟机网络参数
type VMNetworkParams struct {
	VNetName       string   `json:"vnetName"`
	SubnetName     string   `json:"subnetName"`
	PublicIP       bool     `json:"publicIp"`
	SecurityGroups []string `json:"securityGroups"`
}

// VMQueryParams 定义虚拟机查询参数
type VMQueryParams struct {
	// 身份标识
	UserID         string `json:"userId"`
	AccountID      string `json:"accountId"`
	SubscriptionID string `json:"subscriptionId,omitempty"`
	VMID           string `json:"vmId,omitempty"` // 新增VMID字段

	// 基础信息过滤
	Name          string            `json:"name,omitempty"`
	ResourceGroup string            `json:"resourceGroup,omitempty"`
	Location      string            `json:"location,omitempty"`
	Status        string            `json:"status,omitempty"`
	Size          string            `json:"size,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	SyncStatus    string            `json:"syncStatus,omitempty"`

	// 时间范围过滤
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`

	// 分页排序
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	SortBy    string `json:"sortBy,omitempty"`
	SortOrder string `json:"sortOrder,omitempty"`
}

// SyncStats 同步统计信息
type SyncStats struct {
	TotalVMs   int `json:"totalVMs"`   // 同步的总虚拟机数量
	RunningVMs int `json:"runningVMs"` // 运行中的虚拟机数量
	StoppedVMs int `json:"stoppedVMs"` // 已停止的虚拟机数量
}

type UpdateDNSLabelRequest struct {
	DNSLabel string `json:"dnsLabel" binding:"required" example:"my-vm-dns"` // DNS标签
}
