package repository

import (
	"azure-vm-backend/internal/model"
	"context"
	"fmt"
	"time"
)

type VmSizeRepository interface {
	ListVmSizes(ctx context.Context, location string) ([]*model.VmSize, error)
	BatchUpsertVmSizes(ctx context.Context, sizes []*model.VmSize) error
}

type vmSizeRepository struct {
	*Repository
}

func NewVmSizeRepository(repository *Repository) VmSizeRepository {
	return &vmSizeRepository{
		Repository: repository,
	}
}

func (r *vmSizeRepository) ListVmSizes(ctx context.Context, location string) ([]*model.VmSize, error) {
	var sizes []*model.VmSize
	if err := r.DB(ctx).Where("location = ? AND enabled = ?", location, true).Find(&sizes).Error; err != nil {
		return nil, fmt.Errorf("查询规格列表失败: %w", err)
	}
	return sizes, nil
}

func (r *vmSizeRepository) BatchUpsertVmSizes(ctx context.Context, sizes []*model.VmSize) error {
	return r.Transaction(ctx, func(ctx context.Context) error {
		now := time.Now()

		for _, size := range sizes {
			var existing model.VmSize
			err := r.DB(ctx).Where("name = ? AND location = ?", size.Name, size.Location).First(&existing).Error

			if err == nil {
				// 更新现有记录
				existing.Cores = size.Cores
				existing.MemoryGB = size.MemoryGB
				existing.MaxDataDisks = size.MaxDataDisks
				existing.OSDiskSizeGB = size.OSDiskSizeGB
				existing.Category = size.Category
				existing.Family = size.Family
				existing.UpdatedAt = now
				existing.LastSyncAt = now

				if err := r.DB(ctx).Save(&existing).Error; err != nil {
					return fmt.Errorf("更新规格失败: %w", err)
				}
			} else {
				// 插入新记录
				size.CreatedAt = now
				size.UpdatedAt = now
				size.LastSyncAt = now
				size.Enabled = true

				if err := r.DB(ctx).Create(size).Error; err != nil {
					return fmt.Errorf("创建规格失败: %w", err)
				}
			}
		}
		return nil
	})
}
