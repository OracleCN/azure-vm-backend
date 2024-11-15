package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"context"
)

type VmRegionService interface {
	GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error)
}

func NewVmRegionService(
	service *Service,
	vmRegionRepository repository.VmRegionRepository,
) VmRegionService {
	return &vmRegionService{
		Service:            service,
		vmRegionRepository: vmRegionRepository,
	}
}

type vmRegionService struct {
	*Service
	vmRegionRepository repository.VmRegionRepository
}

func (s *vmRegionService) GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error) {
	return s.vmRegionRepository.GetVmRegion(ctx, id)
}
