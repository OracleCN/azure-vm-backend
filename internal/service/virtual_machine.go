package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"context"
)

type VirtualMachineService interface {
	GetVirtualMachine(ctx context.Context, id int64) (*model.VirtualMachine, error)
}

func NewVirtualMachineService(
	service *Service,
	virtualMachineRepository repository.VirtualMachineRepository,
) VirtualMachineService {
	return &virtualMachineService{
		Service:                  service,
		virtualMachineRepository: virtualMachineRepository,
	}
}

type virtualMachineService struct {
	*Service
	virtualMachineRepository repository.VirtualMachineRepository
}

func (s *virtualMachineService) GetVirtualMachine(ctx context.Context, id int64) (*model.VirtualMachine, error) {
	return s.virtualMachineRepository.GetVirtualMachine(ctx, id)
}
