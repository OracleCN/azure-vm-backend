package repository

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/pkg/app"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type AccountsRepository interface {
	GetAccountByEmail(ctx context.Context, email string) (*model.Accounts, error)
	Create(ctx context.Context, account *model.Accounts) error
	GetAccountByUserIdAndEmail(ctx context.Context, userId string, email string) (*model.Accounts, error)
	GetAccountsByUserId(ctx context.Context, userId string, option *app.QueryOption) (*app.ListResult[*model.Accounts], error)
	GetAccountByUserIdAndAccountId(ctx context.Context, userId string, accountId string) (*model.Accounts, error)
	DeleteAccount(ctx context.Context, userId string, accountId string) error
	UpdateAccount(ctx context.Context, userId string, accountId string, updates map[string]interface{}) error
	BatchDeleteAccounts(ctx context.Context, userId string, accountIds []string) (int64, error)
	UpdateVMCount(ctx context.Context, accountID string, vmCount int64) error
	GetAccountsByIDs(ctx context.Context, userId string, accountIds []string) ([]*model.Accounts, error)
	GetNotExistAccountIDs(ctx context.Context, userId string, accountIds []string) ([]string, error)
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

// Create 创建账户
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

// GetAccountsByUserId 获取用户的账户列表，支持分页和搜索
func (r *Repository) GetAccountsByUserId(ctx context.Context, userId string, option *app.QueryOption) (*app.ListResult[*model.Accounts], error) {
	baseQuery := func(db *gorm.DB) *gorm.DB {
		query := db.Model(&model.Accounts{}).Where("user_id = ?", userId)

		// 处理搜索条件
		if search := option.Filters["search"]; search != "" {
			searchTerm := "%" + search + "%"
			query = query.Where(
				db.Where("login_email LIKE ?", searchTerm).
					Or("remark LIKE ?", searchTerm))
		}

		return query
	}

	// 使用通用分页查询
	return app.WithPagination[*model.Accounts](r.db.WithContext(ctx), option, baseQuery)
}

// DeleteAccount 删除账户
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

// UpdateAccount 更新账户信息
func (r *Repository) UpdateAccount(ctx context.Context, userId string, accountId string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&model.Accounts{}).
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

// BatchDeleteAccounts 批量删除账户
func (r *Repository) BatchDeleteAccounts(ctx context.Context, userId string, accountIds []string) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id IN ?", userId, accountIds).
		Delete(&model.Accounts{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetAccountByUserIdAndAccountId 根据用户id和账户id获取账户信息
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

// UpdateVMCount 更新账户的虚拟机数量
func (r *accountsRepository) UpdateVMCount(ctx context.Context, accountID string, vmCount int64) error {
	if accountID == "" {
		return fmt.Errorf("帐户ID不能为空")
	}

	// 开始事务
	tx := r.DB(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("事务开始失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新账户表中的虚拟机数量和最后更新时间
	result := tx.Model(&model.Accounts{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"vm_count":   vmCount,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("更新虚拟机数量失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("帐户未找到: %s", accountID)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交失败: %w", err)
	}

	return nil
}

// GetAccountsByIDs 根据用户ID和账户ID列表获取账户信息
func (r *Repository) GetAccountsByIDs(ctx context.Context, userId string, accountIds []string) ([]*model.Accounts, error) {
	var accounts []*model.Accounts

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id IN ?", userId, accountIds).
		Find(&accounts).Error
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}

	return accounts, nil
}

// GetNotExistAccountIDs 获取不存在的账户ID列表
func (r *Repository) GetNotExistAccountIDs(ctx context.Context, userId string, accountIds []string) ([]string, error) {
	var existingIds []string

	err := r.db.WithContext(ctx).
		Model(&model.Accounts{}).
		Where("user_id = ? AND account_id IN ?", userId, accountIds).
		Pluck("account_id", &existingIds).Error
	if err != nil {
		return nil, fmt.Errorf("查询账户ID失败: %w", err)
	}

	// 找出不存在的ID
	existingMap := make(map[string]bool)
	for _, id := range existingIds {
		existingMap[id] = true
	}

	notExistIds := make([]string, 0)
	for _, id := range accountIds {
		if !existingMap[id] {
			notExistIds = append(notExistIds, id)
		}
	}

	return notExistIds, nil
}
