package repository

import (
	"azure-vm-backend/internal/model"
	"context"
)

type VmSizeRepository interface {
	GetVmSize(ctx context.Context, id int64) (*model.VmSize, error)
}

func NewVmSizeRepository(
	repository *Repository,
) VmSizeRepository {
	return &vmSizeRepository{
		Repository: repository,
	}
}

type vmSizeRepository struct {
	*Repository
}

func (r *vmSizeRepository) GetVmSize(ctx context.Context, id int64) (*model.VmSize, error) {
	var vmSize model.VmSize

	return &vmSize, nil
}
