package service

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/azure"
	"context"
	"time"
)

type VmRegionService interface {
	GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error)
	ListVmRegions(ctx context.Context, enabled *bool) ([]*model.VmRegion, error)
	SyncVmRegions(ctx context.Context, cred *azure.AzureCredential, subscriptionID string) error
	UpdateVmRegion(ctx context.Context, region *model.VmRegion) error
}

type vmRegionService struct {
	*Service
	vmRegionRepository repository.VmRegionRepository
}

func NewVmRegionService(
	service *Service,
	vmRegionRepository repository.VmRegionRepository,
) VmRegionService {
	return &vmRegionService{
		Service:            service,
		vmRegionRepository: vmRegionRepository,
	}
}

// GetVmRegion 获取单个区域信息
func (s *vmRegionService) GetVmRegion(ctx context.Context, id int64) (*model.VmRegion, error) {
	return s.vmRegionRepository.GetVmRegion(ctx, id)
}

// ListVmRegions 获取区域列表
func (s *vmRegionService) ListVmRegions(ctx context.Context, enabled *bool) ([]*model.VmRegion, error) {
	return s.vmRegionRepository.ListVmRegions(ctx, enabled)
}

// SyncVmRegions 同步Azure区域信息
func (s *vmRegionService) SyncVmRegions(ctx context.Context, cred *azure.AzureCredential, subscriptionID string) error {
	// 创建Azure区域获取器
	fetcher := azure.NewRegionFetcher(s.logger.With(), 3, 30*time.Second)

	// 从Azure获取区域信息
	azureRegions, err := fetcher.GetRegions(ctx, cred, subscriptionID)
	if err != nil {
		return err
	}

	// 转换为数据库模型
	var regions []model.VmRegion
	for _, ar := range azureRegions {
		regions = append(regions, model.VmRegion{
			Name:        ar.Name,
			DisplayName: ar.DisplayName,
			Location:    ar.Location,
			Status:      ar.Status,
			Enabled:     true,
			LastSyncAt:  time.Now(),
		})
	}

	// 同步到数据库
	return s.vmRegionRepository.SyncVmRegions(ctx, regions)
}

// UpdateVmRegion 更新区域信息
func (s *vmRegionService) UpdateVmRegion(ctx context.Context, region *model.VmRegion) error {
	return s.vmRegionRepository.UpdateVmRegion(ctx, region)
}
