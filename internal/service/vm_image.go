package service

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/azure"
	"context"
	"fmt"
)

type VmImageService interface {
	GetVmImage(ctx context.Context, id uint) (*model.VmImage, error)
	ListVmImages(ctx context.Context) ([]*model.VmImage, error)
	SyncVmImages(ctx context.Context, userId, accountId, subscriptionId, location string) error
	GetVmImageBySpec(ctx context.Context, publisher, offer, sku string) (*model.VmImage, error)
}

type vmImageService struct {
	*Service
	vmImageRepository       repository.VmImageRepository
	accountsRepository      repository.AccountsRepository
	subscriptionsRepository repository.SubscriptionsRepository
}

func NewVmImageService(
	service *Service,
	vmImageRepository repository.VmImageRepository,
	accountsRepository repository.AccountsRepository,
	subscriptionsRepository repository.SubscriptionsRepository,
) VmImageService {
	return &vmImageService{
		Service:                 service,
		vmImageRepository:       vmImageRepository,
		accountsRepository:      accountsRepository,
		subscriptionsRepository: subscriptionsRepository,
	}
}

func (s *vmImageService) GetVmImage(ctx context.Context, id uint) (*model.VmImage, error) {
	return s.vmImageRepository.GetVmImage(ctx, id)
}

func (s *vmImageService) ListVmImages(ctx context.Context) ([]*model.VmImage, error) {
	return s.vmImageRepository.ListVmImages(ctx)
}

func (s *vmImageService) GetVmImageBySpec(ctx context.Context, publisher, offer, sku string) (*model.VmImage, error) {
	return s.vmImageRepository.GetVmImageBySpec(ctx, publisher, offer, sku)
}

func (s *vmImageService) SyncVmImages(ctx context.Context, userId, accountId, subscriptionId, location string) error {
	// 1. 验证账户权限
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userId, accountId)
	if err != nil {
		return fmt.Errorf("获取账户信息失败: %w", err)
	}
	if account == nil {
		return v1.ErrUnauthorized
	}

	// 2. 验证订阅
	subscription, err := s.subscriptionsRepository.GetSubscription(ctx, accountId, subscriptionId)
	if err != nil {
		return fmt.Errorf("获取订阅信息失败: %w", err)
	}
	if subscription == nil {
		return v1.ErrorAzureNotFound
	}

	// 创建 VMImageFetcher 实例
	fetcher := azure.NewVMImageFetcher(
		subscriptionId,
		&azure.AzureCredential{
			TenantID:     account.Tenant,
			ClientID:     account.AppID,
			ClientSecret: account.PassWord,
		},
		s.logger.With(),
	)

	// 从Azure获取镜像信息
	azureImages, err := fetcher.SyncImages(ctx, location)
	if err != nil {
		return fmt.Errorf("从Azure同步镜像信息失败: %w", err)
	}

	// 转换为数据库模型
	var dbImages []*model.VmImage
	for _, img := range azureImages {
		dbImage := &model.VmImage{
			Publisher:   img.Publisher,
			Offer:       img.Offer,
			Sku:         img.SKU,
			Version:     img.Version,
			OSType:      img.OSType,
			DisplayName: img.DisplayName,
			Description: img.Description,
			Enabled:     true,
		}
		dbImages = append(dbImages, dbImage)
	}

	// 批量更新数据库
	if err := s.vmImageRepository.BatchUpsertVmImages(ctx, dbImages); err != nil {
		return fmt.Errorf("更新数据库镜像信息失败: %w", err)
	}

	return nil
}
