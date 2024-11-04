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
func (r *Repository) UpsertSubscriptions(ctx context.Context, subscriptions []*model.Subscriptions) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(subscriptions) == 0 {
			return nil
		}
		accountID := subscriptions[0].AccountID

		// 1. 获取该账号下所有现有订阅（包括已删除的）
		var existingSubs []model.Subscriptions
		if err := tx.Unscoped().Where("account_id = ?", accountID).Find(&existingSubs).Error; err != nil {
			return err
		}

		// 2. 创建映射以快速查找现有订阅
		existingSubMap := make(map[string]*model.Subscriptions)
		for i := range existingSubs {
			existingSubMap[existingSubs[i].SubscriptionID] = &existingSubs[i]
		}

		// 3. 遍历新订阅数据
		for _, newSub := range subscriptions {
			if existing, ok := existingSubMap[newSub.SubscriptionID]; ok {
				// 3.1 如果订阅已存在，更新它
				updates := map[string]interface{}{
					"display_name":          newSub.DisplayName,
					"state":                 newSub.State,
					"subscription_policies": newSub.SubscriptionPolicies,
					"authorization_source":  newSub.AuthorizationSource,
					"subscription_type":     newSub.SubscriptionType,
					"spending_limit":        newSub.SpendingLimit,
					"start_date":            newSub.StartDate,
					"end_date":              newSub.EndDate,
					"deleted_at":            nil, // 恢复已删除的记录
				}

				if err := tx.Unscoped().Model(existing).Updates(updates).Error; err != nil {
					return err
				}

				// 从映射中删除，剩下的就是需要被删除的订阅
				delete(existingSubMap, newSub.SubscriptionID)
			} else {
				// 3.2 如果是新订阅，直接创建
				if err := tx.Create(newSub).Error; err != nil {
					return err
				}
			}
		}

		// 4. 软删除不再存在的订阅
		var subsToDelete []string
		for subID := range existingSubMap {
			subsToDelete = append(subsToDelete, subID)
		}
		if len(subsToDelete) > 0 {
			if err := tx.Where("account_id = ? AND subscription_id IN ?", accountID, subsToDelete).
				Delete(&model.Subscriptions{}).Error; err != nil {
				return err
			}
		}

		return nil
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
