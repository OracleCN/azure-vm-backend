package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"context"
)

type VmImageService interface {
	GetVmImage(ctx context.Context, id int64) (*model.VmImage, error)
}

func NewVmImageService(
	service *Service,
	vmImageRepository repository.VmImageRepository,
) VmImageService {
	return &vmImageService{
		Service:           service,
		vmImageRepository: vmImageRepository,
	}
}

type vmImageService struct {
	*Service
	vmImageRepository repository.VmImageRepository
}

func (s *vmImageService) GetVmImage(ctx context.Context, id int64) (*model.VmImage, error) {
	return s.vmImageRepository.GetVmImage(ctx, id)
}
