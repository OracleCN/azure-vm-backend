package model

import (
	"time"

	"gorm.io/gorm"
)

// VMSize Azure虚拟机规格信息
type VmSize struct {
	gorm.Model
	Name         string    `gorm:"column:name;type:varchar(64);uniqueIndex;not null" json:"name"`
	DisplayName  string    `gorm:"column:display_name;type:varchar(128)" json:"displayName"`
	CPUCores     int       `gorm:"column:cpu_cores;not null" json:"cpuCores"`
	MemoryGB     float64   `gorm:"column:memory_gb;not null" json:"memoryGb"`
	MaxDataDisks int       `gorm:"column:max_data_disks;not null" json:"maxDataDisks"`
	OSDiskSizeGB int       `gorm:"column:os_disk_size_gb;not null" json:"osDiskSizeGb"`
	Region       string    `gorm:"column:region;type:varchar(64);not null" json:"region"`
	Price        float64   `gorm:"column:price;not null" json:"price"`
	Enabled      bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	LastSyncAt   time.Time `gorm:"column:last_sync_at" json:"lastSyncAt"`
}

func (m *VmSize) TableName() string {
	return "vm_size"
}
