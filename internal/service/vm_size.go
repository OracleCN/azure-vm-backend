package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"context"
)

type VmSizeService interface {
	GetVmSize(ctx context.Context, id int64) (*model.VmSize, error)
}

func NewVmSizeService(
	service *Service,
	vmSizeRepository repository.VmSizeRepository,
) VmSizeService {
	return &vmSizeService{
		Service:          service,
		vmSizeRepository: vmSizeRepository,
	}
}

type vmSizeService struct {
	*Service
	vmSizeRepository repository.VmSizeRepository
}

func (s *vmSizeService) GetVmSize(ctx context.Context, id int64) (*model.VmSize, error) {
	return s.vmSizeRepository.GetVmSize(ctx, id)
}
