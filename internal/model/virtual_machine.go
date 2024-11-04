package model

import (
	"gorm.io/gorm"
	"time"
)

// VirtualMachine 虚拟机模型
type VirtualMachine struct {
	gorm.Model
	VMID           string    `gorm:"column:vm_id;type:varchar(128);uniqueIndex;not null" json:"vmId"`
	AccountID      string    `gorm:"column:account_id;type:varchar(32);index;not null" json:"accountId"`
	SubscriptionID string    `gorm:"column:subscription_id;type:varchar(128);index;not null" json:"subscriptionId"`
	Name           string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	ResourceGroup  string    `gorm:"column:resource_group;type:varchar(128);not null" json:"resourceGroup"`
	Location       string    `gorm:"column:location;type:varchar(64);not null" json:"location"`
	Size           string    `gorm:"column:size;type:varchar(64);not null" json:"size"`
	Status         string    `gorm:"column:status;type:varchar(32);not null;default:Running" json:"status"`
	State          string    `gorm:"column:state;type:varchar(32);not null" json:"state"`
	PrivateIPs     string    `gorm:"column:private_ips;type:text" json:"privateIps"` // JSON string array
	PublicIPs      string    `gorm:"column:public_ips;type:text" json:"publicIps"`   // JSON string array
	OSType         string    `gorm:"column:os_type;type:varchar(32)" json:"osType"`
	OSDiskSize     int       `gorm:"column:os_disk_size;type:integer" json:"osDiskSize"`
	DataDisks      string    `gorm:"column:data_disks;type:text" json:"dataDisks"` // JSON array of disk objects
	Tags           string    `gorm:"column:tags;type:text" json:"tags"`            // JSON object
	SyncStatus     string    `gorm:"column:sync_status;type:varchar(32);not null;default:pending" json:"syncStatus"`
	LastSyncAt     time.Time `gorm:"column:last_sync_at" json:"lastSyncAt"`
}

// TableName 指定表名
func (vm *VirtualMachine) TableName() string {
	return "virtual_machines"
}
