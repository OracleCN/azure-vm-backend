package repository

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"context"
	"errors"
	"gorm.io/gorm"
)

type AccountsRepository interface {
	GetAccountByEmail(ctx context.Context, email string) (*model.Accounts, error)
	Create(ctx context.Context, account *model.Accounts) error
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

// GetAccountByEmail 检查邮箱是否已被使用
func (r *accountsRepository) GetAccountByEmail(ctx context.Context, email string) (*model.Accounts, error) {
	var account model.Accounts

	if err := r.DB(ctx).Where("login_email = ?", email).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 未找到记录，说明邮箱未被使用
			return nil, nil
		}
		// 其他数据库错误
		return nil, err
	}

	// 找到记录，说明邮箱已被使用
	return &account, nil
}
func (r *accountsRepository) Create(ctx context.Context, account *model.Accounts) error {
	// 开启事务
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 再次检查邮箱是否存在（防止并发创建）
		var count int64
		if err := tx.Model(&model.Accounts{}).
			Where("user_id = ? AND login_email = ?", account.UserID, account.LoginEmail).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return v1.ErrAccountEmailDuplicate
		}

		// 2. 创建账号记录
		if err := tx.Create(account).Error; err != nil {
			// 处理唯一索引冲突等数据库错误
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return v1.ErrAccountEmailDuplicate
			}
			return err
		}

		return nil
	})
}
