package repository

import (
	"azure-vm-backend/internal/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type VmImageRepository interface {
	GetVmImage(ctx context.Context, id uint) (*model.VmImage, error)
	ListVmImages(ctx context.Context) ([]*model.VmImage, error)
	BatchUpsertVmImages(ctx context.Context, images []*model.VmImage) error
	GetVmImageBySpec(ctx context.Context, publisher, offer, sku string) (*model.VmImage, error)
}

type vmImageRepository struct {
	*Repository
}

func NewVmImageRepository(repository *Repository) VmImageRepository {
	return &vmImageRepository{
		Repository: repository,
	}
}

func (r *vmImageRepository) GetVmImage(ctx context.Context, id uint) (*model.VmImage, error) {
	var vmImage model.VmImage
	if err := r.DB(ctx).First(&vmImage, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询镜像失败: %w", err)
	}
	return &vmImage, nil
}

func (r *vmImageRepository) ListVmImages(ctx context.Context) ([]*model.VmImage, error) {
	var images []*model.VmImage
	if err := r.DB(ctx).Where("enabled = ?", true).Find(&images).Error; err != nil {
		return nil, fmt.Errorf("查询镜像列表失败: %w", err)
	}
	return images, nil
}

func (r *vmImageRepository) GetVmImageBySpec(ctx context.Context, publisher, offer, sku string) (*model.VmImage, error) {
	var image model.VmImage
	if err := r.DB(ctx).Where("publisher = ? AND offer = ? AND sku = ? AND enabled = ?",
		publisher, offer, sku, true).First(&image).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询镜像失败: %w", err)
	}
	return &image, nil
}

func (r *vmImageRepository) BatchUpsertVmImages(ctx context.Context, images []*model.VmImage) error {
	return r.Transaction(ctx, func(ctx context.Context) error {
		if len(images) == 0 {
			return nil
		}

		now := time.Now()

		// 构建唯一键映射，用于快速查找
		var keys []string
		keyToImage := make(map[string]*model.VmImage)
		for _, img := range images {
			key := fmt.Sprintf("%s:%s:%s", img.Publisher, img.Offer, img.Sku)
			keys = append(keys, key)
			keyToImage[key] = img
		}

		// 查询现有记录
		var existingImages []*model.VmImage
		// 使用 SQLite 的字符串连接语法
		if err := r.DB(ctx).Where("publisher || ':' || offer || ':' || sku IN ?", keys).Find(&existingImages).Error; err != nil {
			return fmt.Errorf("查询现有镜像失败: %w", err)
		}

		// 分类需要更新和插入的记录
		var toUpdate []*model.VmImage
		existingKeys := make(map[string]bool)

		for _, existing := range existingImages {
			key := fmt.Sprintf("%s:%s:%s", existing.Publisher, existing.Offer, existing.Sku)
			existingKeys[key] = true

			if newImage := keyToImage[key]; newImage != nil {
				// 更新现有记录
				existing.Version = newImage.Version
				existing.OSType = newImage.OSType
				existing.DisplayName = newImage.DisplayName
				existing.Description = newImage.Description
				existing.UpdatedAt = now
				existing.LastSyncAt = now
				toUpdate = append(toUpdate, existing)
			}
		}

		// 收集需要插入的记录
		var toInsert []*model.VmImage
		for key, img := range keyToImage {
			if !existingKeys[key] {
				img.CreatedAt = now
				img.UpdatedAt = now
				img.LastSyncAt = now
				img.Enabled = true
				toInsert = append(toInsert, img)
			}
		}

		// 批量插入新记录
		if len(toInsert) > 0 {
			if err := r.DB(ctx).CreateInBatches(toInsert, 100).Error; err != nil {
				return fmt.Errorf("批量插入镜像失败: %w", err)
			}
			r.logger.Info("批量插入镜像成功",
				zap.Int("count", len(toInsert)))
		}

		// 批量更新现有记录
		if len(toUpdate) > 0 {
			for _, img := range toUpdate {
				if err := r.DB(ctx).Save(img).Error; err != nil {
					return fmt.Errorf("更新镜像失败: %w", err)
				}
			}
			r.logger.Info("批量更新镜像成功",
				zap.Int("count", len(toUpdate)))
		}

		return nil
	})
}
