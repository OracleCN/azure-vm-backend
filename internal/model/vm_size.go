package model

import (
	"time"

	"gorm.io/gorm"
)

// VmSize Azure虚拟机规格信息
type VmSize struct {
	gorm.Model
	Name         string    `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Location     string    `gorm:"column:location;type:varchar(64);not null" json:"location"`
	Cores        int       `gorm:"column:cores;not null" json:"cores"`
	MemoryGB     float64   `gorm:"column:memory_gb;not null" json:"memoryGB"`
	MaxDataDisks int       `gorm:"column:max_data_disks;not null" json:"maxDataDisks"`
	OSDiskSizeGB int       `gorm:"column:os_disk_size_gb;not null" json:"osDiskSizeGB"`
	Category     string    `gorm:"column:category;type:varchar(32)" json:"category"`
	Family       string    `gorm:"column:family;type:varchar(32)" json:"family"`
	Enabled      bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	LastSyncAt   time.Time `gorm:"column:last_sync_at" json:"lastSyncAt"`
	PricePerHour float64   `gorm:"column:price_per_hour" json:"pricePerHour"`
	Currency     string    `gorm:"column:currency;type:varchar(16)" json:"currency"`
}

func (m *VmSize) TableName() string {
	return "vm_sizes"
}
