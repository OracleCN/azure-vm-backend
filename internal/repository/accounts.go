package repository

import (
	"azure-vm-backend/internal/model"
	"context"
)

type AccountsRepository interface {
	GetAccounts(ctx context.Context, id int64) (*model.Accounts, error)
}

func NewAccountsRepository(
	repository *Repository,
) AccountsRepository {
	return &accountsRepository{
		Repository: repository,
	}
}

type accountsRepository struct {
	*Repository
}

func (r *accountsRepository) GetAccounts(ctx context.Context, id int64) (*model.Accounts, error) {
	var accounts model.Accounts

	return &accounts, nil
}
