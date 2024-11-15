package model

import (
	"time"

	"gorm.io/gorm"
)

// VMRegion Azure区域信息
type VmRegion struct {
	gorm.Model
	Name        string    `gorm:"column:name;type:varchar(64);uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"column:display_name;type:varchar(128)" json:"displayName"`
	Location    string    `gorm:"column:location;type:varchar(64);not null" json:"location"`
	Status      string    `gorm:"column:status;type:varchar(32);not null" json:"status"`
	Enabled     bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	LastSyncAt  time.Time `gorm:"column:last_sync_at" json:"lastSyncAt"`
}

func (m *VmRegion) TableName() string {
	return "vm_region"
}
