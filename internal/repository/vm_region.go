package repository

import (
	"azure-vm-backend/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type VmRegionRepository interface {
	GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error)
	ListVmRegions(ctx context.Context, enabled *bool) ([]*model.VmRegion, error)
	SyncVmRegions(ctx context.Context, regions []model.VmRegion) error
	UpdateVmRegion(ctx context.Context, region *model.VmRegion) error
}

type vmRegionRepository struct {
	*Repository
}

func NewVmRegionRepository(repository *Repository) VmRegionRepository {
	return &vmRegionRepository{
		Repository: repository,
	}
}

func (r *vmRegionRepository) GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error) {
	var region model.VmRegion
	if err := r.DB(ctx).First(&region, id).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *vmRegionRepository) ListVmRegions(ctx context.Context, enabled *bool) ([]*model.VmRegion, error) {
	var regions []*model.VmRegion
	query := r.DB(ctx)
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	if err := query.Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *vmRegionRepository) SyncVmRegions(ctx context.Context, regions []model.VmRegion) error {
	return r.Transaction(ctx, func(ctx context.Context) error {
		for _, region := range regions {
			var existing model.VmRegion
			err := r.DB(ctx).Where("name = ?", region.Name).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接创建
				region.LastSyncAt = time.Now()
				if err := r.DB(ctx).Create(&region).Error; err != nil {
					return err
				}
			} else if err == nil {
				// 更新现有记录
				existing.DisplayName = region.DisplayName
				existing.Location = region.Location
				existing.Status = region.Status
				existing.LastSyncAt = time.Now()
				if err := r.DB(ctx).Save(&existing).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
}

func (r *vmRegionRepository) UpdateVmRegion(ctx context.Context, region *model.VmRegion) error {
	return r.DB(ctx).Save(region).Error
}
