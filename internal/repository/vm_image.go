package repository

import (
	"azure-vm-backend/internal/model"
	"context"
)

type VmImageRepository interface {
	GetVmImage(ctx context.Context, id int64) (*model.VmImage, error)
}

func NewVmImageRepository(
	repository *Repository,
) VmImageRepository {
	return &vmImageRepository{
		Repository: repository,
	}
}

type vmImageRepository struct {
	*Repository
}

func (r *vmImageRepository) GetVmImage(ctx context.Context, id int64) (*model.VmImage, error) {
	var vmImage model.VmImage

	return &vmImage, nil
}
