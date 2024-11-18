package v1

import (
	"azure-vm-backend/internal/model"
	"time"
)

// VmSizeInfo VM规格信息响应
type VmSizeInfo struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`         // 规格名称，如 Standard_D2s_v3
	Location     string    `json:"location"`     // 区域
	Cores        int       `json:"cores"`        // CPU核心数
	MemoryGB     float64   `json:"memoryGB"`     // 内存大小(GB)
	MaxDataDisks int       `json:"maxDataDisks"` // 最大数据磁盘数
	OSDiskSizeGB int       `json:"osDiskSizeGB"` // OS磁盘大小(GB)
	Category     string    `json:"category"`     // 规格类别
	Family       string    `json:"family"`       // 规格族
	Enabled      bool      `json:"enabled"`      // 是否启用
	LastSyncAt   time.Time `json:"lastSyncAt"`   // 最后同步时间
	CreatedAt    time.Time `json:"createdAt"`    // 创建时间
	UpdatedAt    time.Time `json:"updatedAt"`    // 更新时间
}

// ListVmSizesRequest 获取规格列表请求
type ListVmSizesRequest struct {
	Location string `form:"location" json:"location"` // 区域
}

// ListVmSizesResponse 获取规格列表响应
type ListVmSizesResponse struct {
	Total int64         `json:"total"` // 总数
	Sizes []*VmSizeInfo `json:"sizes"` // 规格列表
}

// SyncVmSizesRequest 同步规格请求
type SyncVmSizesRequest struct {
	AccountID      string `json:"accountId" binding:"required"`      // 账户ID
	SubscriptionID string `json:"subscriptionId" binding:"required"` // 订阅ID
	Location       string `json:"location" binding:"required"`       // 区域
}

// SyncVmSizesResponse 同步规格响应
type SyncVmSizesResponse struct {
	Total    int    `json:"total"`    // 同步的规格总数
	Message  string `json:"message"`  // 同步结果消息
	SyncTime string `json:"syncTime"` // 同步时间
	Location string `json:"location"` // 同步的区域
}

// ToVmSizeInfo 将数据库模型转换为API响应模型
func ToVmSizeInfo(size *model.VmSize) *VmSizeInfo {
	return &VmSizeInfo{
		ID:           size.ID,
		Name:         size.Name,
		Location:     size.Location,
		Cores:        size.Cores,
		MemoryGB:     size.MemoryGB,
		MaxDataDisks: size.MaxDataDisks,
		OSDiskSizeGB: size.OSDiskSizeGB,
		Category:     size.Category,
		Family:       size.Family,
		Enabled:      size.Enabled,
		LastSyncAt:   size.LastSyncAt,
		CreatedAt:    size.CreatedAt,
		UpdatedAt:    size.UpdatedAt,
	}
}

// ToListVmSizesResponse 将数据库结果转换为列表响应
func ToListVmSizesResponse(sizes []*model.VmSize) *ListVmSizesResponse {
	var sizeInfos []*VmSizeInfo
	for _, size := range sizes {
		sizeInfos = append(sizeInfos, ToVmSizeInfo(size))
	}

	return &ListVmSizesResponse{
		Total: int64(len(sizes)),
		Sizes: sizeInfos,
	}
}
