package service

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/app"
	"azure-vm-backend/pkg/azure"
	"azure-vm-backend/pkg/log"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type VirtualMachineService interface {
	// GetVM 查询操作
	GetVM(ctx context.Context, userID string, accountID string, vmID string) (*model.VirtualMachine, error)
	ListVMs(ctx context.Context, params *v1.VMQueryParams) (*app.ListResult[*model.VirtualMachine], error)
	ListVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) ([]*model.VirtualMachine, error)

	// SyncVMs 同步操作
	SyncVMs(ctx context.Context, userID, accountID string) error
	SyncVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) error

	// CreateVM 虚拟机操作 - 这些接口暂时不实现
	CreateVM(ctx context.Context, userID, accountID string, params *v1.VMCreateParams) (*model.VirtualMachine, error)
	DeleteVM(ctx context.Context, userID, accountID, vmID string) error
	StartVM(ctx context.Context, userID, accountID, vmID string) error
	StopVM(ctx context.Context, userID, accountID, vmID string) error
	RestartVM(ctx context.Context, userID, accountID, vmID string) error
}

func convertTags(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	jsonBytes, err := json.Marshal(tags)
	if err != nil {
		// 如果无法序列化，返回空字符串
		return ""
	}
	return string(jsonBytes)
}
func NewVirtualMachineService(
	service *Service,
	virtualMachineRepository repository.VirtualMachineRepository,
) VirtualMachineService {
	return &virtualMachineService{
		Service:                  service,
		virtualMachineRepository: virtualMachineRepository,
	}
}

type virtualMachineService struct {
	*Service
	virtualMachineRepository repository.VirtualMachineRepository
	accountsRepository       repository.AccountsRepository
	subscriptionsRepository  repository.SubscriptionsRepository
	logger                   *log.Logger
}

// GetVM 获取单个虚拟机详细信息
func (s *virtualMachineService) GetVM(ctx context.Context, userID, accountID, vmID string) (*model.VirtualMachine, error) {
	// 检查用户是否有权限访问该账号
	if err := s.checkAccountAccess(ctx, userID, accountID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// 从数据库获取虚拟机信息
	vm, err := s.virtualMachineRepository.GetByID(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	// 验证虚拟机所属账号
	if vm.AccountID != accountID {
		return nil, fmt.Errorf("VM does not belong to the specified account")
	}

	return vm, nil
}

// ListVMs 获取指定账号下的所有虚拟机
func (s *virtualMachineService) ListVMs(ctx context.Context, params *v1.VMQueryParams) (*app.ListResult[*model.VirtualMachine], error) {
	// 构建查询选项
	queryOpt := &app.QueryOption{
		Pagination: app.Pagination{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
		Filters:   make(map[string]string),
	}

	// 设置默认值
	if queryOpt.Page <= 0 {
		queryOpt.Page = 1
	}
	if queryOpt.PageSize <= 0 {
		queryOpt.PageSize = 10
	}
	if queryOpt.SortBy == "" {
		queryOpt.SortBy = "created_at"
	}
	if queryOpt.SortOrder == "" {
		queryOpt.SortOrder = "desc"
	}

	// 构建VM查询选项
	vmOpts := repository.QueryVMsOptions{
		AccountID:      params.AccountID,
		SubscriptionID: params.SubscriptionID,
		Query:          queryOpt,
		ExtraFilters:   make(map[string]string),
	}

	// 添加基础过滤条件
	if params.Name != "" {
		vmOpts.ExtraFilters["name"] = params.Name
	}
	if params.ResourceGroup != "" {
		vmOpts.ExtraFilters["resource_group"] = params.ResourceGroup
	}
	if params.Location != "" {
		vmOpts.ExtraFilters["location"] = params.Location
	}
	if params.Status != "" {
		vmOpts.ExtraFilters["status"] = params.Status
	}
	if params.Size != "" {
		vmOpts.ExtraFilters["size"] = params.Size
	}
	if params.SyncStatus != "" {
		vmOpts.ExtraFilters["sync_status"] = params.SyncStatus
	}

	// 添加标签过滤
	if len(params.Tags) > 0 {
		for k, v := range params.Tags {
			vmOpts.ExtraFilters[fmt.Sprintf("tag_%s", k)] = v
		}
	}

	// 添加时间范围过滤
	if params.StartTime != nil {
		vmOpts.ExtraFilters["start_time"] = params.StartTime.Format(time.RFC3339)
	}
	if params.EndTime != nil {
		vmOpts.ExtraFilters["end_time"] = params.EndTime.Format(time.RFC3339)
	}

	// 执行查询
	return s.virtualMachineRepository.ListVMs(ctx, vmOpts)
}

// SyncVMs 同步指定账号下的所有虚拟机信息
func (s *virtualMachineService) SyncVMs(ctx context.Context, userID, accountID string) error {
	// 检查用户是否有权限访问该账号
	if err := s.checkAccountAccess(ctx, userID, accountID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	// 获取账号凭据
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account credentials: %w", err)
	}

	// 创建Azure凭据
	credentials := &azure.Credentials{
		TenantID:     account.Tenant,
		ClientID:     account.AppID,
		ClientSecret: account.PassWord,
	}

	// 创建VM获取器
	vmFetcher := azure.NewVMFetcher(credentials, s.logger.With(
		zap.String("accountId", accountID),
		zap.String("userId", userID),
	), 5*time.Minute)

	// 获取最新的VM信息
	vms, err := vmFetcher.FetchVMDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch VMs from Azure: %w", err)
	}

	// 将Azure VM信息转换为数据库模型
	var dbVMs []*model.VirtualMachine
	for _, vm := range vms {
		dbVM := &model.VirtualMachine{
			AccountID:      accountID,
			VMID:           vm.ID,
			SubscriptionID: vm.SubscriptionID,
			Name:           vm.Name,
			ResourceGroup:  vm.ResourceGroup,
			Location:       vm.Location,
			Size:           vm.Size,
			Status:         vm.State,
			Tags:           convertTags(vm.Tags),
			SyncStatus:     "synced",
			LastSyncAt:     time.Now(),
			// 设置其他必要字段
		}
		dbVMs = append(dbVMs, dbVM)
	}

	// 批量更新数据库
	if err := s.virtualMachineRepository.BatchUpsert(ctx, dbVMs); err != nil {
		return fmt.Errorf("failed to update VMs in database: %w", err)
	}

	return nil
}

// ListVMsBySubscription 获取指定订阅下的所有虚拟机
func (s *virtualMachineService) ListVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) ([]*model.VirtualMachine, error) {
	// 检查用户是否有权限访问该账号
	if err := s.checkAccountAccess(ctx, userID, accountID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// 检查订阅是否属于该账号
	if err := s.checkSubscriptionAccess(ctx, accountID, subscriptionID); err != nil {
		return nil, fmt.Errorf("subscription access denied: %w", err)
	}

	// 创建基础查询选项
	queryOpt := &app.QueryOption{
		Pagination: app.Pagination{
			Page:     1,
			PageSize: 100,
		},
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// 查询虚拟机列表
	result, err := s.virtualMachineRepository.ListBySubscriptionID(ctx, subscriptionID, queryOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs by subscription: %w", err)
	}

	return result.Items, nil
}

// SyncVMsBySubscription 同步指定订阅下的所有虚拟机信息
func (s *virtualMachineService) SyncVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) error {
	// 检查用户是否有权限访问该账号
	if err := s.checkAccountAccess(ctx, userID, accountID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	// 检查订阅是否属于该账号
	if err := s.checkSubscriptionAccess(ctx, accountID, subscriptionID); err != nil {
		return fmt.Errorf("subscription access denied: %w", err)
	}

	// 获取账号凭据
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account credentials: %w", err)
	}

	// 创建Azure凭据
	credentials := &azure.Credentials{
		TenantID:     account.Tenant,
		ClientID:     account.AppID,
		ClientSecret: account.PassWord,
	}

	// 创建VM获取器
	vmFetcher := azure.NewVMFetcher(credentials, s.logger.With(
		zap.String("accountId", accountID),
		zap.String("subscriptionId", subscriptionID),
		zap.String("userId", userID),
	), 5*time.Minute)

	// 获取最新的VM信息
	vms, err := vmFetcher.FetchVMDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch VMs from Azure: %w", err)
	}

	// 过滤出指定订阅的VM
	var subscriptionVMs []*model.VirtualMachine
	for _, vm := range vms {
		if vm.SubscriptionID == subscriptionID {
			dbVM := &model.VirtualMachine{
				AccountID:      accountID,
				VMID:           vm.ID,
				SubscriptionID: vm.SubscriptionID,
				Name:           vm.Name,
				ResourceGroup:  vm.ResourceGroup,
				Location:       vm.Location,
				Size:           vm.Size,
				Status:         vm.State,
				Tags:           convertTags(vm.Tags),
				SyncStatus:     "synced",
				LastSyncAt:     time.Now(),
				// 设置其他必要字段
			}
			subscriptionVMs = append(subscriptionVMs, dbVM)
		}
	}

	// 批量更新数据库
	if err := s.virtualMachineRepository.BatchUpsert(ctx, subscriptionVMs); err != nil {
		return fmt.Errorf("failed to update VMs in database: %w", err)
	}

	return nil
}

// checkAccountAccess 检查用户是否有权限访问指定账号
func (s *virtualMachineService) checkAccountAccess(ctx context.Context, userID, accountID string) error {
	account, err := s.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userID, accountID)
	if err != nil {
		return err
	}

	// 检查账号是否属于该用户
	if account.UserID != userID {
		return fmt.Errorf("user does not have access to this account")
	}

	return nil
}

// checkSubscriptionAccess 检查订阅是否属于指定账号
func (s *virtualMachineService) checkSubscriptionAccess(ctx context.Context, accountID, subscriptionID string) error {
	subscription, err := s.subscriptionsRepository.GetSubscription(ctx, accountID, subscriptionID)
	if err != nil {
		return err
	}

	// 检查订阅是否属于该账号
	if subscription.AccountID != accountID {
		return fmt.Errorf("subscription does not belong to this account")
	}

	return nil
}

// CreateVM 创建新的虚拟机 - 暂不实现
func (s *virtualMachineService) CreateVM(ctx context.Context, userID, accountID string, params *v1.VMCreateParams) (*model.VirtualMachine, error) {
	// TODO: 预留创建虚拟机接口
	return nil, v1.ErrNotImplemented
}

// DeleteVM 删除虚拟机 - 暂不实现
func (s *virtualMachineService) DeleteVM(ctx context.Context, userID, accountID, vmID string) error {
	// TODO: 预留删除虚拟机接口
	return v1.ErrNotImplemented
}

// StartVM 启动虚拟机 - 暂不实现
func (s *virtualMachineService) StartVM(ctx context.Context, userID, accountID, vmID string) error {
	// TODO: 预留启动虚拟机接口
	return v1.ErrNotImplemented
}

// StopVM 停止虚拟机 - 暂不实现
func (s *virtualMachineService) StopVM(ctx context.Context, userID, accountID, vmID string) error {
	// TODO: 预留停止虚拟机接口
	return v1.ErrNotImplemented
}

// RestartVM 重启虚拟机 - 暂不实现
func (s *virtualMachineService) RestartVM(ctx context.Context, userID, accountID, vmID string) error {
	// TODO: 预留重启虚拟机接口
	return v1.ErrNotImplemented
}
