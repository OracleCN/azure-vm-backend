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
	GetAccountByUserIdAndEmail(ctx context.Context, userId string, email string) (*model.Accounts, error)
	GetAccountsByUserId(ctx context.Context, userId string) ([]*model.Accounts, error)
	GetAccountByUserIdAndAccountId(ctx context.Context, userId string, accountId string) (*model.Accounts, error)
	DeleteAccount(ctx context.Context, userId string, accountId string) error
	UpdateAccount(ctx context.Context, userId string, accountId string, updates map[string]interface{}) error
	BatchDeleteAccounts(ctx context.Context, userId string, accountIds []string) (int64, error)
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

// GetAccountByUserIdAndEmail 根据用户id和邮箱获取账户信息
func (r *Repository) GetAccountByUserIdAndEmail(ctx context.Context, userId string, email string) (*model.Accounts, error) {
	var account model.Accounts

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND login_email = ?", userId, email).
		First(&account)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &account, nil
}

func (r *Repository) GetAccountsByUserId(ctx context.Context, userId string) ([]*model.Accounts, error) {
	var accounts []*model.Accounts

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Find(&accounts)

	if result.Error != nil {
		return nil, result.Error
	}

	return accounts, nil
}

func (r *Repository) DeleteAccount(ctx context.Context, userId string, accountId string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id = ?", userId, accountId).
		Delete(&model.Accounts{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *Repository) UpdateAccount(ctx context.Context, userId string, accountId string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id = ?", userId, accountId).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
func (r *Repository) BatchDeleteAccounts(ctx context.Context, userId string, accountIds []string) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id IN ?", userId, accountIds).
		Delete(&model.Accounts{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
func (r *Repository) GetAccountByUserIdAndAccountId(ctx context.Context, userId string, accountId string) (*model.Accounts, error) {
	var account model.Accounts

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id = ?", userId, accountId).
		First(&account)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &account, nil
}
