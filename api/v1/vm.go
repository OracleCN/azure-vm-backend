package v1

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
