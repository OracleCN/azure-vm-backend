package repository

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/pkg/app"
	"context"
	"gorm.io/gorm"
	"time"
)

// QueryVMsOptions VM查询选项
type QueryVMsOptions struct {
	AccountID      string            // 账号ID
	SubscriptionID string            // 订阅ID
	Query          *app.QueryOption  // 通用查询选项(分页、排序等)
	ExtraFilters   map[string]string // 额外的过滤条件
}

type VirtualMachineRepository interface {
	Create(ctx context.Context, vm *model.VirtualMachine) error
	GetByID(ctx context.Context, vmID string) (*model.VirtualMachine, error)
	Update(ctx context.Context, vm *model.VirtualMachine) error
	Delete(ctx context.Context, vmID string) error

	// ListVMs 查询操作
	ListVMs(ctx context.Context, opts QueryVMsOptions) (*app.ListResult[*model.VirtualMachine], error)
	ListBySubscriptionID(ctx context.Context, subscriptionID string, query *app.QueryOption) (*app.ListResult[*model.VirtualMachine], error)
	ListByAccountID(ctx context.Context, accountID string, query *app.QueryOption) (*app.ListResult[*model.VirtualMachine], error)
	ListByAccountAndSubscription(ctx context.Context, accountID string, subscriptionID string, query *app.QueryOption) (*app.ListResult[*model.VirtualMachine], error)
	// UpdateSyncStatus 同步相关操作
	UpdateSyncStatus(ctx context.Context, vmID string, status string, syncTime time.Time) error
	BatchUpsert(ctx context.Context, vms []*model.VirtualMachine) error

	// UpdateStatus 状态相关操作
	UpdateStatus(ctx context.Context, vmID string, status string) error
}

func NewVirtualMachineRepository(
	repository *Repository,
) VirtualMachineRepository {
	return &virtualMachineRepository{
		Repository: repository,
	}
}

type virtualMachineRepository struct {
	*Repository
}

// Create 创建虚拟机记录
func (r *virtualMachineRepository) Create(ctx context.Context, vm *model.VirtualMachine) error {
	// TODO: 实现虚拟机创建逻辑
	// 1. 验证必要字段
	// 2. 检查重复记录
	// 3. 创建记录
	return nil
}

// ListVMs 统一的虚拟机查询方法
func (r *virtualMachineRepository) ListVMs(ctx context.Context, opts QueryVMsOptions) (*app.ListResult[*model.VirtualMachine], error) {
	opts.Query = app.ValidateAndFillQueryOption(opts.Query)

	baseQuery := func(db *gorm.DB) *gorm.DB {
		q := db.Model(&model.VirtualMachine{})

		// 添加基本查询条件
		if opts.AccountID != "" {
			q = q.Where("account_id = ?", opts.AccountID)
		}
		if opts.SubscriptionID != "" {
			q = q.Where("subscription_id = ?", opts.SubscriptionID)
		}

		// 添加通用过滤条件
		if opts.Query.Filters != nil {
			for field, value := range opts.Query.Filters {
				if value != "" {
					switch field {
					case "status":
						q = q.Where("status = ?", value)
					case "location":
						q = q.Where("location = ?", value)
					case "resource_group":
						q = q.Where("resource_group = ?", value)
					case "size":
						q = q.Where("size = ?", value)
					case "sync_status":
						q = q.Where("sync_status = ?", value)
					}
				}
			}
		}

		// 添加额外的过滤条件
		if opts.ExtraFilters != nil {
			for field, value := range opts.ExtraFilters {
				if value != "" {
					switch field {
					case "os_type":
						q = q.Where("os_type = ?", value)
					case "name_like":
						q = q.Where("name LIKE ?", "%"+value+"%")
					case "tag":
						q = q.Where("tags LIKE ?", "%"+value+"%")
						// 可以根据需要添加更多过滤条件
					}
				}
			}
		}

		return q
	}

	return app.WithPagination[*model.VirtualMachine](r.DB(ctx), opts.Query, baseQuery)
}

// GetByID 根据ID获取虚拟机信息
func (r *virtualMachineRepository) GetByID(ctx context.Context, vmID string) (*model.VirtualMachine, error) {
	var vm model.VirtualMachine

	result := r.DB(ctx).Where("vm_id = ?", vmID).First(&vm)
	if result.Error != nil {
		return nil, result.Error
	}

	return &vm, nil
}

// Update 更新虚拟机信息
func (r *virtualMachineRepository) Update(ctx context.Context, vm *model.VirtualMachine) error {
	// TODO: 实现虚拟机更新逻辑
	// 1. 验证更新字段
	// 2. 检查记录是否存在
	// 3. 执行更新操作
	return nil
}

// Delete 删除虚拟机记录
func (r *virtualMachineRepository) Delete(ctx context.Context, vmID string) error {
	// TODO: 实现虚拟机删除逻辑
	// 1. 验证ID
	// 2. 检查记录是否存在
	// 3. 执行软删除操作
	return nil
}

// UpdateSyncStatus 更新虚拟机同步状态
func (r *virtualMachineRepository) UpdateSyncStatus(ctx context.Context, vmID string, status string, syncTime time.Time) error {
	// TODO: 实现更新同步状态逻辑
	// 1. 验证状态值是否合法
	// 2. 更新同步状态和时间
	// 3. 记录状态变更历史
	return nil
}

// BatchUpsert 批量更新或插入虚拟机记录
func (r *virtualMachineRepository) BatchUpsert(ctx context.Context, vms []*model.VirtualMachine) error {
	// TODO: 实现批量更新插入逻辑
	// 1. 验证虚拟机记录列表
	// 2. 批量检查已存在记录
	// 3. 执行批量插入或更新操作
	return nil
}

// UpdateStatus 更新虚拟机状态
func (r *virtualMachineRepository) UpdateStatus(ctx context.Context, vmID string, status string) error {
	// TODO: 实现更新虚拟机状态逻辑
	// 1. 验证状态值是否合法
	// 2. 更新虚拟机状态
	// 3. 记录状态变更历史
	return nil
}

// ListByAccountID 获取指定账号的所有虚拟机
func (r *virtualMachineRepository) ListByAccountID(ctx context.Context, accountID string, query *app.QueryOption) (*app.ListResult[*model.VirtualMachine], error) {
	return r.ListVMs(ctx, QueryVMsOptions{
		AccountID: accountID,
		Query:     query,
	})
}

// ListBySubscriptionID 获取指定订阅的所有虚拟机
func (r *virtualMachineRepository) ListBySubscriptionID(ctx context.Context, subscriptionID string, query *app.QueryOption) (*app.ListResult[*model.VirtualMachine], error) {
	return r.ListVMs(ctx, QueryVMsOptions{
		SubscriptionID: subscriptionID,
		Query:          query,
	})
}

// ListByAccountAndSubscription 获取指定账号和订阅的所有虚拟机
func (r *virtualMachineRepository) ListByAccountAndSubscription(ctx context.Context, accountID string, subscriptionID string, query *app.QueryOption) (*app.ListResult[*model.VirtualMachine], error) {
	return r.ListVMs(ctx, QueryVMsOptions{
		AccountID:      accountID,
		SubscriptionID: subscriptionID,
		Query:          query,
	})
}
