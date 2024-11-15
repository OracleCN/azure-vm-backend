package repository

import (
	"azure-vm-backend/internal/model"
	"context"
)

type VmRegionRepository interface {
	GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error)
}

func NewVmRegionRepository(
	repository *Repository,
) VmRegionRepository {
	return &vmRegionRepository{
		Repository: repository,
	}
}

type vmRegionRepository struct {
	*Repository
}

func (r *vmRegionRepository) GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error) {
	var vmRegion model.VmRegion

	return &vmRegion, nil
}
