package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"context"
)

type AccountsService interface {
	GetAccounts(ctx context.Context, id int64) (*model.Accounts, error)
}

func NewAccountsService(
	service *Service,
	accountsRepository repository.AccountsRepository,
) AccountsService {
	return &accountsService{
		Service:            service,
		accountsRepository: accountsRepository,
	}
}

type accountsService struct {
	*Service
	accountsRepository repository.AccountsRepository
}

func (s *accountsService) GetAccounts(ctx context.Context, id int64) (*model.Accounts, error) {
	return s.accountsRepository.GetAccounts(ctx, id)
}
