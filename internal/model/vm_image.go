package model

import (
	"gorm.io/gorm"
	"time"
)

type VmImage struct {
	gorm.Model
	ImageID      string    `gorm:"column:image_id;type:varchar(128);uniqueIndex;not null" json:"imageId"` // 镜像ID
	Name         string    `gorm:"column:name;type:varchar(128);not null" json:"name"`                    // 镜像名称
	RegionID     string    `gorm:"column:region_id;type:varchar(64);index;not null" json:"regionId"`      // 所属区域ID
	Publisher    string    `gorm:"column:publisher;type:varchar(128);not null" json:"publisher"`          // 发布者
	Offer        string    `gorm:"column:offer;type:varchar(128);not null" json:"offer"`                  // 产品
	SKU          string    `gorm:"column:sku;type:varchar(128);not null" json:"sku"`                      // SKU
	Version      string    `gorm:"column:version;type:varchar(64);not null" json:"version"`               // 版本
	OSType       string    `gorm:"column:os_type;type:varchar(32);not null" json:"osType"`                // 操作系统类型
	Category     string    `gorm:"column:category;type:varchar(32);not null" json:"category"`             // 分类(如 Recommended)
	Available    bool      `gorm:"column:available;not null;default:true" json:"available"`               // 是否可用
	Description  string    `gorm:"column:description;type:text" json:"description"`                       // 描述
	LastSyncAt   time.Time `gorm:"column:last_sync_at;type:datetime;not null" json:"lastSyncAt"`          // 最后同步时间
	Requirements string    `gorm:"column:requirements;type:text" json:"requirements"`                     // 系统要求(JSON)
	Features     string    `gorm:"column:features;type:text" json:"features"`                             // 特性(JSON)
	Metadata     string    `gorm:"column:metadata;type:text" json:"metadata"`                             // 元数据(JSON)
}

func (m *VmImage) TableName() string {
	return "vm_image"
}
