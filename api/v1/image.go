package v1

import (
	"azure-vm-backend/internal/model"
	"time"
)

// ImageInfo 镜像信息响应
type ImageInfo struct {
	ID          uint      `json:"id"`
	Publisher   string    `json:"publisher"`   // 发布者
	Offer       string    `json:"offer"`       // 产品
	Sku         string    `json:"sku"`         // SKU
	Version     string    `json:"version"`     // 版本
	OSType      string    `json:"osType"`      // 操作系统类型
	DisplayName string    `json:"displayName"` // 显示名称
	Description string    `json:"description"` // 描述
	Enabled     bool      `json:"enabled"`     // 是否启用
	LastSyncAt  time.Time `json:"lastSyncAt"`  // 最后同步时间
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`   // 更新时间
}

// ListImagesRequest 获取镜像列表请求
type ListImagesRequest struct {
	Page     int    `form:"page" json:"page"`         // 页码
	PageSize int    `form:"pageSize" json:"pageSize"` // 每页数量
	Search   string `form:"search" json:"search"`     // 搜索关键词
	OSType   string `form:"osType" json:"osType"`     // 操作系统类型过滤
}

// ListImagesResponse 获取镜像列表响应
type ListImagesResponse struct {
	Total  int64        `json:"total"`  // 总数
	Images []*ImageInfo `json:"images"` // 镜像列表
}

// SyncImagesRequest 同步镜像请求
type SyncImagesRequest struct {
	AccountID      string `json:"accountId" binding:"required"`      // 账户ID
	SubscriptionID string `json:"subscriptionId" binding:"required"` // 订阅ID
	Location       string `json:"location" binding:"required"`       // 区域
}

// SyncImagesResponse 同步镜像响应
type SyncImagesResponse struct {
	Total    int    `json:"total"`    // 同步的镜像总数
	Message  string `json:"message"`  // 同步结果消息
	SyncTime string `json:"syncTime"` // 同步时间
	Location string `json:"location"` // 同步的区域
}

// GetImageResponse 获取镜像详情响应
type GetImageResponse struct {
	ImageInfo
}

// ToImageInfo 将数据库模型转换为API响应模型
func ToImageInfo(image *model.VmImage) *ImageInfo {
	return &ImageInfo{
		ID:          image.ID,
		Publisher:   image.Publisher,
		Offer:       image.Offer,
		Sku:         image.Sku,
		Version:     image.Version,
		OSType:      image.OSType,
		DisplayName: image.DisplayName,
		Description: image.Description,
		Enabled:     image.Enabled,
		LastSyncAt:  image.LastSyncAt,
		CreatedAt:   image.CreatedAt,
		UpdatedAt:   image.UpdatedAt,
	}
}

// ToListImagesResponse 将数据库结果转换为列表响应
func ToListImagesResponse(images []*model.VmImage, total int64, page, pageSize int) *ListImagesResponse {
	var imageInfos []*ImageInfo
	for _, image := range images {
		imageInfos = append(imageInfos, ToImageInfo(image))
	}

	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}

	return &ListImagesResponse{
		Total:  total,
		Images: imageInfos,
	}
}
