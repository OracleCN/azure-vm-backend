package v1

import (
	"azure-vm-backend/internal/model"
	"time"
)

// VmRegionResp 区域信息响应
type VmRegionResp struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`        // 区域名称
	DisplayName string    `json:"displayName"` // 显示名称
	Location    string    `json:"location"`    // 地理位置
	Status      string    `json:"status"`      // 状态
	Enabled     bool      `json:"enabled"`     // 是否启用
	LastSyncAt  time.Time `json:"lastSyncAt"`  // 最后同步时间
}

// ListVmRegionReq 获取区域列表请求
type ListVmRegionReq struct {
	Enabled *bool `form:"enabled" json:"enabled"` // 是否只返回启用的区域
}

// ListVmRegionResp 区域列表响应
type ListVmRegionResp struct {
	Total   int64           `json:"total"`
	Regions []*VmRegionResp `json:"regions"`
}

// SyncVmRegionReq 同步区域信息请求
type SyncVmRegionReq struct {
	TenantID       string `json:"tenantId" binding:"required"`
	ClientID       string `json:"clientId" binding:"required"`
	ClientSecret   string `json:"clientSecret" binding:"required"`
	SubscriptionID string `json:"subscriptionId" binding:"required"`
}

// UpdateVmRegionReq 更新区域信息请求
type UpdateVmRegionReq struct {
	Enabled bool `json:"enabled" binding:"required"` // 是否启用
}

// 转换函数
func ToVmRegionResp(region *model.VmRegion) *VmRegionResp {
	if region == nil {
		return nil
	}
	return &VmRegionResp{
		ID:          region.ID,
		Name:        region.Name,
		DisplayName: region.DisplayName,
		Location:    region.Location,
		Status:      region.Status,
		Enabled:     region.Enabled,
		LastSyncAt:  region.LastSyncAt,
	}
}

func ToVmRegionListResp(regions []*model.VmRegion) *ListVmRegionResp {
	resp := &ListVmRegionResp{
		Total:   int64(len(regions)),
		Regions: make([]*VmRegionResp, 0, len(regions)),
	}
	for _, region := range regions {
		resp.Regions = append(resp.Regions, ToVmRegionResp(region))
	}
	return resp
}
