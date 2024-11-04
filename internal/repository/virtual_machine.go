package repository

import (
	"azure-vm-backend/internal/model"
	"context"
)

type VirtualMachineRepository interface {
	GetVirtualMachine(ctx context.Context, id int64) (*model.VirtualMachine, error)
}

func NewVirtualMachineRepository(
	repository *Repository,
) VirtualMachineRepository {
	return &virtualMachineRepository{
		Repository: repository,
	}
}

type virtualMachineRepository struct {
	*Repository
}

func (r *virtualMachineRepository) GetVirtualMachine(ctx context.Context, id int64) (*model.VirtualMachine, error) {
	var virtualMachine model.VirtualMachine

	return &virtualMachine, nil
}
