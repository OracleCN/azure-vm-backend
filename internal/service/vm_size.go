package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/azure"
	"context"
	"fmt"
)

type VmSizeService interface {
	ListVmSizes(ctx context.Context, location string) ([]*model.VmSize, error)
	SyncVmSizes(ctx context.Context, userId, accountId, subscriptionId, location string) error
}

type vmSizeService struct {
	*Service
	vmSizeRepository        repository.VmSizeRepository
	accountsRepository      repository.AccountsRepository
	subscriptionsRepository repository.SubscriptionsRepository
}

func NewVmSizeService(
	service *Service,
	vmSizeRepository repository.VmSizeRepository,
	accountsRepository repository.AccountsRepository,
	subscriptionsRepository repository.SubscriptionsRepository,
) VmSizeService {
	return &vmSizeService{
		Service:                 service,
		vmSizeRepository:        vmSizeRepository,
		accountsRepository:      accountsRepository,
		subscriptionsRepository: subscriptionsRepository,
	}
}

func (s *vmSizeService) ListVmSizes(ctx context.Context, location string) ([]*model.VmSize, error) {
	return s.vmSizeRepository.ListVmSizes(ctx, location)
}

func (s *vmSizeService) SyncVmSizes(ctx context.Context, userId, accountId, subscriptionId, location string) error {
	// 验证账户权限
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userId, accountId)
	if err != nil {
		return fmt.Errorf("获取账户信息失败: %w", err)
	}
	if account == nil {
		return fmt.Errorf("账户不存在")
	}

	// 验证订阅
	subscription, err := s.subscriptionsRepository.GetSubscription(ctx, accountId, subscriptionId)
	if err != nil {
		return fmt.Errorf("获取订阅信息失败: %w", err)
	}
	if subscription == nil {
		return fmt.Errorf("订阅不存在")
	}

	// 创建 Azure 客户端
	fetcher := azure.NewVMSizeFetcher(
		subscriptionId,
		&azure.AzureCredential{
			TenantID:     account.Tenant,
			ClientID:     account.AppID,
			ClientSecret: account.PassWord,
		},
		s.logger.With(),
	)

	// 获取规格信息
	sizes, err := fetcher.ListSizes(ctx, location)
	if err != nil {
		return fmt.Errorf("获取规格列表失败: %w", err)
	}

	// 转换为数据库模型
	var dbSizes []*model.VmSize
	for _, size := range sizes {
		dbSize := &model.VmSize{
			Name:         size.Name,
			Location:     location,
			Cores:        size.Cores,
			MemoryGB:     size.MemoryGB,
			MaxDataDisks: size.MaxDataDisks,
			OSDiskSizeGB: size.OSDiskSizeGB,
			Category:     size.Category,
			Family:       size.Family,
			Enabled:      true,
		}
		dbSizes = append(dbSizes, dbSize)
	}

	// 更新数据库
	if err := s.vmSizeRepository.BatchUpsertVmSizes(ctx, dbSizes); err != nil {
		return fmt.Errorf("更新数据库失败: %w", err)
	}

	return nil
}
