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
	"strconv"
	"strings"
	"time"
)

type VirtualMachineService interface {
	// GetVM 查询操作
	GetVM(ctx context.Context, userID string, accountID string, vmID string) (*model.VirtualMachine, error)
	ListVMs(ctx context.Context, params *v1.VMQueryParams) (*app.ListResult[*model.VirtualMachine], error)
	ListVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) ([]*model.VirtualMachine, error)

	// SyncVMs 同步操作
	SyncVMs(ctx context.Context, userID, accountID string) (*v1.SyncStats, error)
	SyncVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) error

	// CreateVM 虚拟机操作 - 这些接口暂时不实现
	CreateVM(ctx context.Context, userID, accountID string, params *v1.VMCreateParams) (*model.VirtualMachine, error)
	DeleteVM(ctx context.Context, userID, accountID, vmID string) error
	StartVM(ctx context.Context, userID, accountID, vmID string) error
	StopVM(ctx context.Context, userID, accountID, vmID string) error
	RestartVM(ctx context.Context, userID, accountID, vmID string) error
	// UpdateDNSLabel 更新DNS标签
	UpdateDNSLabel(ctx context.Context, userId string, accountId string, vmId string, dnsLabel string) error
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
	accountsRepository repository.AccountsRepository, // 添加账号仓储
	subscriptionsRepository repository.SubscriptionsRepository, // 添加订阅仓储
	logger *log.Logger, // 添加日志器
) VirtualMachineService {
	return &virtualMachineService{
		Service:                  service,
		virtualMachineRepository: virtualMachineRepository,
		accountsRepository:       accountsRepository,
		subscriptionsRepository:  subscriptionsRepository,
		logger:                   logger,
	}
}

type virtualMachineService struct {
	*Service
	virtualMachineRepository repository.VirtualMachineRepository
	accountsRepository       repository.AccountsRepository
	subscriptionsRepository  repository.SubscriptionsRepository
	logger                   *log.Logger
}

// syncVMsHelper 同步虚拟机的辅助结构体
type syncVMsHelper struct {
	service     *virtualMachineService
	ctx         context.Context
	userID      string
	accountID   string
	credentials *azure.Credentials
	logger      *zap.Logger
}

// newSyncVMsHelper 创建同步辅助结构体
func newSyncVMsHelper(service *virtualMachineService, ctx context.Context, userID, accountID string) (*syncVMsHelper, error) {
	// 获取账号凭据
	account, err := service.accountsRepository.GetAccountByUserIdAndAccountId(ctx, userID, accountID)
	if err != nil {
		return nil, fmt.Errorf("获取帐户凭据失败: %w", err)
	}

	// 创建Azure凭据
	credentials := &azure.Credentials{
		TenantID:     account.Tenant,
		ClientID:     account.AppID,
		ClientSecret: account.PassWord,
	}

	// 创建日志记录器
	logger := service.logger.With(
		zap.String("accountId", accountID),
		zap.String("userId", userID),
	)

	return &syncVMsHelper{
		service:     service,
		ctx:         ctx,
		userID:      userID,
		accountID:   accountID,
		credentials: credentials,
		logger:      logger,
	}, nil
}

// convertVMToModel 将Azure VM转换为数据库模型
func (h *syncVMsHelper) convertVMToModel(vm azure.VMDetails) (*model.VirtualMachine, error) {
	// 转换网络信息为JSON字符串
	networkInfo := map[string]interface{}{
		"privateIPs": vm.PrivateIPs,
		"publicIPs":  vm.PublicIPs,
	}
	_, err := json.Marshal(networkInfo)
	if err != nil {
		return nil, fmt.Errorf("未能调用网络信息: %w", err)
	}

	// 转换磁盘信息为JSON字符串
	diskInfo := map[string]interface{}{
		"osDiskSize": vm.OSDiskSize,
		"dataDisks":  vm.DataDisks,
	}
	diskJSON, err := json.Marshal(diskInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal disk info: %w", err)
	}

	// 创建数据库VM记录
	return &model.VirtualMachine{
		// 基础标识信息
		AccountID:      h.accountID,
		VMID:           vm.ID,
		SubscriptionID: vm.SubscriptionID,
		Name:           vm.Name,
		ResourceGroup:  vm.ResourceGroup,

		// 资源信息
		Location:   vm.Location,
		Size:       vm.Size,
		Status:     vm.State,
		OSType:     vm.OSType,
		OSImage:    vm.OSImage,
		Core:       vm.NumberOfCores,
		Memory:     vm.MemoryInGB,
		DnsAlias:   vm.DnsAlias,
		State:      vm.State,
		PowerState: vm.PowerState,

		// 网络配置
		PrivateIPs:   strings.Join(vm.PrivateIPs, ","),
		PublicIPs:    strings.Join(vm.PublicIPs, ","),
		PublicIPName: vm.PublicIPName,

		// 磁盘配置
		DataDisks:  string(diskJSON),
		OSDiskSize: int(vm.OSDiskSize),

		// 元数据
		Tags: convertTags(vm.Tags),

		// 同步状态
		SyncStatus:  "synced",
		LastSyncAt:  time.Now(),
		CreatedTime: vm.CreatedTime,
	}, nil
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

// SyncVMs 同步指定账号下的所有虚拟机信息
func (s *virtualMachineService) SyncVMs(ctx context.Context, userID, accountID string) (*v1.SyncStats, error) {
	stats := &v1.SyncStats{}
	// 检查用户是否有权限访问该账号
	if err := s.checkAccountAccess(ctx, userID, accountID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// 创建同步辅助结构体
	helper, err := newSyncVMsHelper(s, ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	// 创建VM获取器
	vmFetcher := azure.NewVMFetcher(helper.credentials, helper.logger, 5*time.Minute)

	// 获取最新的VM信息
	vms, err := vmFetcher.FetchVMDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("从 Azure 获取虚拟机失败: %w", err)
	}

	// 转换所有VM为数据库模型
	var dbVMs []*model.VirtualMachine
	for _, vm := range vms {
		dbVM, err := helper.convertVMToModel(vm)
		if err != nil {
			helper.logger.Error("转换虚拟机失败",
				zap.String("vmId", vm.ID),
				zap.Error(err))
			continue
		}
		switch dbVM.PowerState {
		case "running":
			stats.RunningVMs++
		case "deallocated":
			stats.StoppedVMs++
		}
		dbVMs = append(dbVMs, dbVM)
	}
	stats.TotalVMs = len(vms)
	// 批量更新数据库
	if err := s.virtualMachineRepository.BatchUpsert(ctx, dbVMs); err != nil {
		return nil, fmt.Errorf("更新数据库中的虚拟机失败: %w", err)
	}

	// 更新账户表中的虚拟机数量
	if err := s.accountsRepository.UpdateVMCount(ctx, accountID, int64(len(vms))); err != nil {
		s.logger.Error("更新账户中的虚拟机数量失败",
			zap.String("accountID", accountID),
			zap.String("vmCount", strconv.FormatInt(int64(len(vms)), 10)),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("成功更新账户虚拟机数量",
		zap.String("accountID", accountID),
		zap.Int64("vmCount", int64(len(vms))))
	// 返回同步成功多少台虚拟机，运行中多少台，已停止多少台
	s.logger.Info("成功同步虚拟机信息",
		zap.String("accountID", accountID),
		zap.Int("totalVMs", stats.TotalVMs),
		zap.Int("runningVMs", stats.RunningVMs),
		zap.Int("stoppedVMs", stats.StoppedVMs))
	return stats, nil
}

func (s *virtualMachineService) SyncVMsBySubscription(ctx context.Context, userID, accountID, subscriptionID string) error {
	// 检查用户是否有权限访问该账号
	if err := s.checkAccountAccess(ctx, userID, accountID); err != nil {
		return fmt.Errorf("拒绝访问: %w", err)
	}

	// 检查订阅是否属于该账号
	if err := s.checkSubscriptionAccess(ctx, accountID, subscriptionID); err != nil {
		return fmt.Errorf("拒绝订阅访问: %w", err)
	}

	// 创建同步辅助结构体
	helper, err := newSyncVMsHelper(s, ctx, userID, accountID)
	if err != nil {
		return err
	}

	// 创建VM获取器
	vmFetcher := azure.NewVMFetcher(helper.credentials, helper.logger.With(
		zap.String("subscriptionId", subscriptionID),
	), 5*time.Minute)

	// 获取最新的VM信息
	vms, err := vmFetcher.FetchVMDetails(ctx)
	if err != nil {
		return fmt.Errorf("从 Azure 获取虚拟机失败: %w", err)
	}

	// 过滤并转换指定订阅的VM
	var subscriptionVMs []*model.VirtualMachine
	for _, vm := range vms {
		if vm.SubscriptionID == subscriptionID {
			dbVM, err := helper.convertVMToModel(vm)
			if err != nil {
				helper.logger.Error("转换虚拟机失败",
					zap.String("vmId", vm.ID),
					zap.Error(err))
				continue
			}
			subscriptionVMs = append(subscriptionVMs, dbVM)
		}
	}

	// 批量更新数据库
	if err := s.virtualMachineRepository.BatchUpsert(ctx, subscriptionVMs); err != nil {
		return fmt.Errorf("更新数据库中的虚拟机失败: %w", err)
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
		return fmt.Errorf("用户无权访问此帐户")
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
		return fmt.Errorf("订阅不属于此帐户")
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

func (s *virtualMachineService) UpdateDNSLabel(ctx context.Context, userId string, accountId string, vmId string, dnsLabel string) error {
	// 1. 验证用户权限和账户
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

	// 2. 获取虚拟机信息
	vm, err := s.virtualMachineRepository.GetByID(ctx, vmId)
	if err != nil {
		s.logger.Error("获取虚拟机信息失败",
			zap.Error(err),
			zap.String("vmId", vmId),
		)
		return v1.ErrInternalServerError
	}

	if vm == nil {
		return v1.ErrorAzureNotFound
	}

	// 3. 验证虚拟机属于指定账户
	if vm.AccountID != accountId {
		return v1.ErrUnauthorized
	}

	// 4. 创建Azure凭据
	creds := &azure.Credentials{
		TenantID:     account.Tenant,
		ClientID:     account.AppID,
		ClientSecret: account.PassWord,
		DisplayName:  account.DisplayName,
	}

	// 5. 更新Azure云上的DNS标签
	fetcher := azure.NewVMFetcher(creds, s.logger.With(), 30*time.Second)
	fqdn, err := fetcher.SetVMDNSLabel(
		ctx,
		vm.SubscriptionID,
		vm.ResourceGroup,
		vm.PublicIPName,
		dnsLabel,
	)
	if err != nil {
		s.logger.Error("更新Azure DNS标签失败",
			zap.Error(err),
			zap.String("vmId", vmId),
			zap.String("dnsLabel", dnsLabel),
		)
		return v1.ErrInternalServerError
	}

	// 6. 更新本地数据库中的DNS记录
	err = s.virtualMachineRepository.UpdateDNSLabel(ctx, vmId, fqdn)
	if err != nil {
		s.logger.Error("更新本地DNS记录失败",
			zap.Error(err),
			zap.String("vmId", vmId),
			zap.String("dnsLabel", dnsLabel),
		)
		return v1.ErrInternalServerError
	}

	s.logger.Info("DNS标签更新成功",
		zap.String("vmId", vmId),
		zap.String("dnsLabel", dnsLabel),
		zap.String("fqdn", fqdn),
	)

	return nil
}
