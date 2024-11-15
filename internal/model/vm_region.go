package model

import (
	"gorm.io/gorm"
	"time"
)

// VmRegion Azure虚拟机区域信息
type VmRegion struct {
	gorm.Model
	RegionID      string    `gorm:"column:region_id;type:varchar(64);uniqueIndex;not null" json:"regionId"` // 区域ID
	DisplayName   string    `gorm:"column:display_name;type:varchar(128);not null" json:"displayName"`      // 显示名称
	Name          string    `gorm:"column:name;type:varchar(64);not null" json:"name"`                      // 区域名称(如 eastus)
	Location      string    `gorm:"column:location;type:varchar(64);not null" json:"location"`              // 地理位置
	RegionType    string    `gorm:"column:region_type;type:varchar(32);not null" json:"regionType"`         // 区域类型
	Category      string    `gorm:"column:category;type:varchar(32);not null" json:"category"`              // 区域分类(如 Recommended)
	Available     bool      `gorm:"column:available;not null;default:true" json:"available"`                // 是否可用
	Enabled       bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`                    // 是否启用
	LastSyncAt    time.Time `gorm:"column:last_sync_at;type:datetime;not null" json:"lastSyncAt"`           // 最后同步时间
	PairedRegion  string    `gorm:"column:paired_region;type:varchar(64)" json:"pairedRegion"`              // 配对区域
	ResourceTypes string    `gorm:"column:resource_types;type:text" json:"resourceTypes"`                   // 支持的资源类型(JSON)
	Metadata      string    `gorm:"column:metadata;type:text" json:"metadata"`                              // 元数据(JSON)
}

func (m *VmRegion) TableName() string {
	return "vm_region"
}
