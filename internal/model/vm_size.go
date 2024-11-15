package model

import (
	"gorm.io/gorm"
	"time"
)

type VmSize struct {
	gorm.Model
	SizeID                string    `gorm:"column:size_id;type:varchar(64);uniqueIndex;not null" json:"sizeId"` // 规格ID
	Name                  string    `gorm:"column:name;type:varchar(64);not null" json:"name"`                  // 规格名称
	RegionID              string    `gorm:"column:region_id;type:varchar(64);index;not null" json:"regionId"`   // 所属区域ID
	Category              string    `gorm:"column:category;type:varchar(32);not null" json:"category"`          // 分类(如 General Purpose)
	Family                string    `gorm:"column:family;type:varchar(32);not null" json:"family"`              // 系列
	NumberOfCores         int32     `gorm:"column:number_of_cores;not null" json:"numberOfCores"`               // CPU核心数
	MemoryInGB            float32   `gorm:"column:memory_in_gb;not null" json:"memoryInGB"`                     // 内存大小(GB)
	MaxDataDisks          int32     `gorm:"column:max_data_disks;not null" json:"maxDataDisks"`                 // 最大数据盘数量
	OSDiskSizeInGB        int32     `gorm:"column:os_disk_size_in_gb;not null" json:"osDiskSizeInGB"`           // 系统盘大小(GB)
	Available             bool      `gorm:"column:available;not null;default:true" json:"available"`            // 是否可用
	Enabled               bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`                // 是否启用
	LastSyncAt            time.Time `gorm:"column:last_sync_at;type:datetime;not null" json:"lastSyncAt"`       // 最后同步时间
	NetworkBandwidth      string    `gorm:"column:network_bandwidth;type:varchar(32)" json:"networkBandwidth"`  // 网络带宽
	TemporaryDiskSizeInGB int32     `gorm:"column:temporary_disk_size_in_gb" json:"temporaryDiskSizeInGB"`      // 临时磁盘大小(GB)
	PricePerHour          float64   `gorm:"column:price_per_hour;type:decimal(10,4)" json:"pricePerHour"`       // 每小时价格
	AcceleratedNetworking bool      `gorm:"column:accelerated_networking" json:"acceleratedNetworking"`         // 是否支持加速网络
	Capabilities          string    `gorm:"column:capabilities;type:text" json:"capabilities"`                  // 特性(JSON)
	Metadata              string    `gorm:"column:metadata;type:text" json:"metadata"`                          // 元数据(JSON)
}

func (m *VmSize) TableName() string {
	return "vm_size"
}
