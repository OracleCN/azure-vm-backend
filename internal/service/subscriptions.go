package service

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/azure"
	"context"
	"go.uber.org/zap"
	"time"
)

// SubscriptionsService 订阅服务接口
type SubscriptionsService interface {
	// GetSubscriptions 获取指定账号的所有订阅信息
	GetSubscriptions(ctx context.Context, userId, accountId string) ([]*model.Subscriptions, error)

	// GetSubscription 获取指定订阅的详细信息
	GetSubscription(ctx context.Context, userId, accountId, subscriptionId string) (*model.Subscriptions, error)

	// SyncSubscriptions 同步指定账号的订阅信息
	SyncSubscriptions(ctx context.Context, userId, accountId string) (int, error)

	// DeleteSubscriptions 删除指定账号的所有订阅信息
	DeleteSubscriptions(ctx context.Context, userId, accountId string) error
}

func NewSubscriptionsService(
	service *Service,
	subscriptionsRepository repository.SubscriptionsRepository,
	accountsRepository repository.AccountsRepository,
) SubscriptionsService {
	return &subscriptionsService{
		Service:                service,
		subscriptionRepository: subscriptionsRepository,
		accountsRepository:     accountsRepository,
	}
}

type subscriptionsService struct {
	*Service
	subscriptionRepository repository.SubscriptionsRepository
	accountsRepository     repository.AccountsRepository
}

// GetSubscriptions 获取指定账号的所有订阅信息
func (s *subscriptionsService) GetSubscriptions(ctx context.Context, userId, accountId string) ([]*model.Subscriptions, error) {
	// 1. 验证账户是否存在且属于该用户
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userId, accountId)
	if err != nil {
		s.logger.Error("获取账户信息失败",
			zap.Error(err),
			zap.String("userId", userId),
			zap.String("accountId", accountId),
		)
		return nil, v1.ErrInternalServerError
	}

	if account == nil {
		return nil, v1.ErrorAzureNotFound
	}

	// 2. 获取订阅信息
	subs, err := s.subscriptionRepository.GetSubscriptionsByAccountId(ctx, accountId)
	if err != nil {
		s.logger.Error("获取订阅信息失败",
			zap.Error(err),
			zap.String("accountId", accountId),
		)
		return nil, v1.ErrInternalServerError
	}

	return subs, nil
}

// GetSubscription 获取指定订阅的详细信息
func (s *subscriptionsService) GetSubscription(ctx context.Context, userId, accountId, subscriptionId string) (*model.Subscriptions, error) {
	// 1. 验证账户是否存在且属于该用户
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userId, accountId)
	if err != nil {
		s.logger.Error("获取账户信息失败",
			zap.Error(err),
			zap.String("userId", userId),
			zap.String("accountId", accountId),
		)
		return nil, v1.ErrInternalServerError
	}

	if account == nil {
		return nil, v1.ErrAccountError
	}

	// 2. 获取订阅信息
	sub, err := s.subscriptionRepository.GetSubscription(ctx, accountId, subscriptionId)
	if err != nil {
		s.logger.Error("获取订阅信息失败",
			zap.Error(err),
			zap.String("accountId", accountId),
			zap.String("subscriptionId", subscriptionId),
		)
		return nil, v1.ErrInternalServerError
	}

	if sub == nil {
		return nil, v1.ErrSubscriptionNotFound
	}

	return sub, nil
}

// SyncSubscriptions 同步指定账号的订阅信息
func (s *subscriptionsService) SyncSubscriptions(ctx context.Context, userId, accountId string) (int, error) {
	// 1. 验证账户是否存在且属于该用户
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userId, accountId)
	if err != nil {
		s.logger.Error("获取账户信息失败",
			zap.Error(err),
			zap.String("userId", userId),
			zap.String("accountId", accountId),
		)
		return 0, v1.ErrInternalServerError
	}

	if account == nil {
		return 0, v1.ErrAccountError
	}

	// 2. 创建Azure凭据
	creds := &azure.Credentials{
		TenantID:     account.Tenant,
		ClientID:     account.AppID,
		ClientSecret: account.PassWord,
		DisplayName:  account.DisplayName,
	}

	// 3. 从Azure获取订阅信息
	fetcher := azure.NewFetcher(creds, s.logger.With(), 30*time.Second)
	azureSubs, err := fetcher.FetchSubscriptionDetails(ctx)
	if err != nil {
		s.logger.Error("获取Azure订阅信息失败",
			zap.Error(err),
			zap.String("accountId", accountId),
		)
		// 更新账户状态为错误
		if updateErr := s.accountsRepository.UpdateAccount(ctx, userId, accountId, map[string]interface{}{
			"subscription_status": "error",
		}); updateErr != nil {
			s.logger.Error("更新账户状态失败",
				zap.Error(updateErr),
				zap.String("accountId", accountId),
			)
		}
		return 0, v1.ErrInternalServerError
	}

	// 4. 转换并保存数据
	var subscriptions []*model.Subscriptions
	for _, azureSub := range azureSubs {
		sub := &model.Subscriptions{}
		if err := sub.FromAzureSubscription(accountId, &azureSub); err != nil {
			s.logger.Error("转换订阅信息失败",
				zap.Error(err),
				zap.String("accountId", accountId),
				zap.String("subscriptionId", azureSub.SubscriptionID),
			)
			continue
		}
		subscriptions = append(subscriptions, sub)
	}

	// 5. 保存到数据库
	if err := s.subscriptionRepository.UpsertSubscriptions(ctx, subscriptions); err != nil {
		s.logger.Error("保存订阅信息失败",
			zap.Error(err),
			zap.String("accountId", accountId),
		)
		return 0, v1.ErrInternalServerError
	}

	// 6. 更新账户状态为正常
	if err := s.accountsRepository.UpdateAccount(ctx, userId, accountId, map[string]interface{}{
		"subscription_status": "normal",
	}); err != nil {
		s.logger.Error("更新账户状态失败",
			zap.Error(err),
			zap.String("accountId", accountId),
		)
	}
	// 返回 同步成功多少个订阅
	return int(int64(len(subscriptions))), nil
}

// DeleteSubscriptions 删除指定账号的所有订阅信息
func (s *subscriptionsService) DeleteSubscriptions(ctx context.Context, userId, accountId string) error {
	// 1. 验证账户是否存在且属于该用户
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userId, accountId)
	if err != nil {
		s.logger.Error("获取账户信息失败",
			zap.Error(err),
			zap.String("userId", userId),
			zap.String("accountId", accountId),
		)
		return v1.ErrInternalServerError
	}

	if account == nil {
		return v1.ErrAccountError
	}

	// 2. 删除订阅信息
	if err := s.subscriptionRepository.DeleteSubscriptionsByAccountId(ctx, accountId); err != nil {
		s.logger.Error("删除订阅信息失败",
			zap.Error(err),
			zap.String("accountId", accountId),
		)
		return v1.ErrInternalServerError
	}

	return nil
}
