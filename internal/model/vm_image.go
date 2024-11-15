package model

import (
	"time"

	"gorm.io/gorm"
)

// VmImage Azure虚拟机镜像信息
type VmImage struct {
	gorm.Model
	Publisher   string    `gorm:"column:publisher;type:varchar(128);not null" json:"publisher"`
	Offer       string    `gorm:"column:offer;type:varchar(128);not null" json:"offer"`
	Sku         string    `gorm:"column:sku;type:varchar(128);not null" json:"sku"`
	Version     string    `gorm:"column:version;type:varchar(64);not null" json:"version"`
	OSType      string    `gorm:"column:os_type;type:varchar(32);not null" json:"osType"`
	DisplayName string    `gorm:"column:display_name;type:varchar(256)" json:"displayName"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Enabled     bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	LastSyncAt  time.Time `gorm:"column:last_sync_at" json:"lastSyncAt"`
}

func (m *VmImage) TableName() string {
	return "vm_images"
}
