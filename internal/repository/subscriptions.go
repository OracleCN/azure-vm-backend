package repository

import (
	"azure-vm-backend/internal/model"
	"context"
	"errors"
	"gorm.io/gorm"
)

type SubscriptionsRepository interface {
	// UpsertSubscriptions 批量更新或插入订阅信息
	UpsertSubscriptions(ctx context.Context, subs []*model.Subscriptions) error
	// GetSubscriptionsByAccountId 获取账号下的所有订阅
	GetSubscriptionsByAccountId(ctx context.Context, accountId string) ([]*model.Subscriptions, error)
	// GetSubscription 获取指定的订阅信息
	GetSubscription(ctx context.Context, accountId, subscriptionId string) (*model.Subscriptions, error)
	// DeleteSubscriptionsByAccountId 删除账号下的所有订阅
	DeleteSubscriptionsByAccountId(ctx context.Context, accountId string) error
}

func NewSubscriptionsRepository(
	repository *Repository,
) SubscriptionsRepository {
	return &subscriptionsRepository{
		Repository: repository,
	}
}

type subscriptionsRepository struct {
	*Repository
}

// UpsertSubscriptions 批量更新或插入订阅信息
func (r *subscriptionsRepository) UpsertSubscriptions(ctx context.Context, subs []*model.Subscriptions) error {
	if len(subs) == 0 {
		return nil
	}

	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 获取所有要更新的account_id
		var accountIDs []string
		accountIDMap := make(map[string]bool)
		for _, sub := range subs {
			if !accountIDMap[sub.AccountID] {
				accountIDs = append(accountIDs, sub.AccountID)
				accountIDMap[sub.AccountID] = true
			}
		}

		// 2. 软删除这些账号下的所有现有订阅
		if err := tx.Where("account_id IN ?", accountIDs).Delete(&model.Subscriptions{}).Error; err != nil {
			return err
		}

		// 3. 批量插入新数据
		return tx.Create(&subs).Error
	})
}

// GetSubscriptionsByAccountId 获取账号下的所有订阅
func (r *subscriptionsRepository) GetSubscriptionsByAccountId(ctx context.Context, accountId string) ([]*model.Subscriptions, error) {
	var subs []*model.Subscriptions
	err := r.DB(ctx).Where("account_id = ?", accountId).Find(&subs).Error
	return subs, err
}

// GetSubscription 获取指定的订阅信息
func (r *subscriptionsRepository) GetSubscription(ctx context.Context, accountId, subscriptionId string) (*model.Subscriptions, error) {
	var sub model.Subscriptions
	err := r.DB(ctx).Where("account_id = ? AND subscription_id = ?", accountId, subscriptionId).First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// DeleteSubscriptionsByAccountId 删除账号下的所有订阅
func (r *subscriptionsRepository) DeleteSubscriptionsByAccountId(ctx context.Context, accountId string) error {
	return r.DB(ctx).Where("account_id = ?", accountId).Delete(&model.Subscriptions{}).Error
}
