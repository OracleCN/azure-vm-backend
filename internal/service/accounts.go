package service

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/app"
	"azure-vm-backend/pkg/azure"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"strings"
	"time"
)

type AccountsService interface {
	// CreateAccount 创建账户
	CreateAccount(ctx context.Context, userId string, req *v1.CreateAccountReq) (string, error)
	GetAccount(ctx context.Context, userId string, loginMail string) (*model.Accounts, error)
	GetAccountList(ctx context.Context, userId string, option *app.QueryOption) (*app.ListResult[*model.Accounts], error)
	DeleteAccount(ctx context.Context, userId string, accountIds []string) error
	UpdateAccount(ctx context.Context, userId string, accountId string, req *v1.UpdateAccountReq) error
	SyncAccounts(ctx context.Context, userId string, accountIds []string) (*v1.SyncAccountResp, error)
}

type accountsService struct {
	*Service
	accountsRepo          repository.AccountsRepository
	subscriptionsService  SubscriptionsService  // 添加订阅服务
	virtualMachineService VirtualMachineService // 添加虚拟机服务
}

func NewAccountsService(
	service *Service,
	accountsRepo repository.AccountsRepository,
	subscriptionsService SubscriptionsService,
	virtualMachineService VirtualMachineService,
) AccountsService {
	return &accountsService{
		Service:               service,
		accountsRepo:          accountsRepo,
		subscriptionsService:  subscriptionsService,
		virtualMachineService: virtualMachineService,
	}
}

func (s *accountsService) CreateAccount(ctx context.Context, userId string, req *v1.CreateAccountReq) (string, error) {
	// 判断 邮箱是否重复出现在 账号表内
	existingAccount, err := s.accountsRepo.GetAccountByEmail(ctx, req.LoginEmail)
	if err != nil {
		s.logger.Error("未检查电子邮件是否存在",
			zap.Error(err),
			zap.String("email", req.LoginEmail),
		)
		return "", v1.ErrAccountError
	}
	if existingAccount != nil {
		s.logger.Warn("已使用的电子邮件",
			zap.String("email", req.LoginEmail),
			zap.String("existing_user_id", existingAccount.UserID),
		)
		return "", v1.ErrAccountEmailDuplicate
	}
	// 2. 验证 Azure 凭据
	validator := azure.NewValidator(60 * time.Second)
	result := validator.ValidateWithContext(ctx, azure.Credentials{
		TenantID:     req.Tenant,      // tenant
		ClientID:     req.AppID,       // appId
		ClientSecret: req.PassWord,    // password
		DisplayName:  req.DisplayName, // displayName
	})

	if !result.Valid {
		s.logger.Error("azure验证失败",
			zap.Error(result.Error),
			zap.String("message", result.Message),
			zap.Time("validated_at", result.ValidatedAt),
			zap.String("display_name", req.DisplayName),
		)
		return "", fmt.Errorf("azure验证失败: %s", result.Message)
	}
	// 3. 创建账号记录
	account := &model.Accounts{
		AccountID:          uuid.New().String(),
		UserID:             userId,
		LoginEmail:         req.LoginEmail,
		LoginPassword:      req.LoginPassword,
		Remark:             req.Remark,
		AppID:              req.AppID,
		PassWord:           req.PassWord,
		Tenant:             req.Tenant,
		DisplayName:        req.DisplayName,
		VmCount:            req.VmCount,
		SubscriptionStatus: "normal",
	}

	if err := s.accountsRepo.Create(ctx, account); err != nil {
		s.logger.Error("failed to create account",
			zap.Error(err),
			zap.String("user_id", userId),
			zap.String("login_email", req.LoginEmail),
		)
		return "", fmt.Errorf("创建账户失败: %w", err)
	}

	s.logger.Info("帐户已成功创建",
		zap.String("account_id", account.AccountID),
		zap.String("user_id", userId),
		zap.String("login_email", req.LoginEmail),
	)

	return account.AccountID, nil
}

// GetAccount 获取某个azure账户的azure账户信息
func (s *accountsService) GetAccount(ctx context.Context, userId string, loginMail string) (*model.Accounts, error) {
	// 判断该用户是否拥有这个azure账号
	email, err := s.accountsRepo.GetAccountByUserIdAndEmail(ctx, userId, loginMail)
	if err != nil {
		s.logger.Error("获取Azure账户失败", zap.Error(err), zap.String("email", loginMail))
		return nil, v1.ErrInternalServerError
	}
	// email 为空
	if email == nil {
		return nil, v1.ErrorAzureNotFound
	}
	// 如果有则返回 这个azure的账户信息
	return email, nil
}

func (s *accountsService) GetAccountList(ctx context.Context, userId string, option *app.QueryOption) (*app.ListResult[*model.Accounts], error) {
	result, err := s.accountsRepo.GetAccountsByUserId(ctx, userId, option)
	if err != nil {
		s.logger.Error("获取用户账户列表失败",
			zap.Error(err),
			zap.String("user_id", userId),
		)
		return nil, v1.ErrInternalServerError
	}

	s.logger.Info("成功获取用户账户列表",
		zap.String("user_id", userId),
		zap.Int("count", len(result.Items)),
	)
	return result, nil
}

func (s *accountsService) DeleteAccount(ctx context.Context, userId string, accountIds []string) error {
	// 参数校验
	if len(accountIds) == 0 {
		return fmt.Errorf("账户ID列表不能为空")
	}

	// 执行批量删除
	deletedCount, err := s.accountsRepo.BatchDeleteAccounts(ctx, userId, accountIds)
	if err != nil {
		s.logger.Error("删除账户失败",
			zap.Error(err),
			zap.String("user_id", userId),
			zap.Strings("account_ids", accountIds),
		)
		return v1.ErrInternalServerError
	}

	// 处理未找到记录的情况
	if deletedCount == 0 {
		s.logger.Warn("未找到要删除的账户",
			zap.String("user_id", userId),
			zap.Strings("account_ids", accountIds),
		)
		return v1.ErrorAzureNotFound
	}

	s.logger.Info("成功删除账户",
		zap.String("user_id", userId),
		zap.Strings("account_ids", accountIds),
		zap.Int64("deleted_count", deletedCount),
	)
	return nil
}

func (s *accountsService) UpdateAccount(ctx context.Context, userId string, accountId string, req *v1.UpdateAccountReq) error {
	// 初始化更新字段map
	updates := make(map[string]interface{})

	// 辅助函数：如果值不为空则添加到更新map中
	addIfNotEmpty := func(dbField string, value string) {
		if value != "" {
			updates[dbField] = value
		}
	}

	// 如果要更新邮箱，先检查邮箱是否已被使用
	if req.LoginEmail != "" {
		existingAccount, err := s.accountsRepo.GetAccountByEmail(ctx, req.LoginEmail)
		if err != nil {
			s.logger.Error("检查邮箱是否存在失败", zap.Error(err))
			return v1.ErrInternalServerError
		}
		if existingAccount != nil && existingAccount.AccountID != accountId {
			return v1.ErrAccountEmailDuplicate
		}
	}

	// 统一处理所有字段更新
	addIfNotEmpty("login_email", req.LoginEmail)
	addIfNotEmpty("login_password", req.LoginPassword)
	addIfNotEmpty("remark", req.Remark)
	addIfNotEmpty("app_id", req.AppID)
	addIfNotEmpty("password", req.PassWord)
	addIfNotEmpty("tenant", req.Tenant)
	addIfNotEmpty("display_name", req.DisplayName)

	// 如果有Azure凭据相关的更新，需要验证新凭据
	if req.AppID != "" || req.PassWord != "" || req.Tenant != "" {
		validator := azure.NewValidator(60 * time.Second)
		result := validator.ValidateWithContext(ctx, azure.Credentials{
			TenantID:     req.Tenant,
			ClientID:     req.AppID,
			ClientSecret: req.PassWord,
			DisplayName:  req.DisplayName,
		})

		if !result.Valid {
			s.logger.Error("azure验证失败",
				zap.Error(result.Error),
				zap.String("message", result.Message),
			)
			return fmt.Errorf("azure验证失败: %s", result.Message)
		}
	}

	// 执行更新
	err := s.accountsRepo.UpdateAccount(ctx, userId, accountId, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ErrorAzureNotFound
		}
		s.logger.Error("更新账户失败",
			zap.Error(err),
			zap.String("user_id", userId),
			zap.String("account_id", accountId),
		)
		return v1.ErrInternalServerError
	}

	s.logger.Info("成功更新账户",
		zap.String("user_id", userId),
		zap.String("account_id", accountId),
	)
	return nil
}

// SyncAccounts 同步多个Azure账户信息
func (s *accountsService) SyncAccounts(ctx context.Context, userId string, accountIds []string) (*v1.SyncAccountResp, error) {
	result := &v1.SyncAccountResp{
		SuccessAccounts: make([]v1.SyncAccountResult, 0),
		FailedAccounts:  make([]v1.SyncAccountResult, 0),
	}

	// 检查哪些账户不存在
	notExistIds, err := s.accountsRepo.GetNotExistAccountIDs(ctx, userId, accountIds)
	if err != nil {
		s.logger.Error("检查账户存在性失败",
			zap.Error(err),
			zap.String("userId", userId))
		return nil, fmt.Errorf("检查账户失败: %w", err)
	}

	// 记录不存在的账户
	for _, accountId := range notExistIds {
		result.FailedAccounts = append(result.FailedAccounts, v1.SyncAccountResult{
			AccountID: accountId,
			Message:   "账户不存在",
		})
	}

	// 获取存在的账户信息
	accounts, err := s.accountsRepo.GetAccountsByIDs(ctx, userId, accountIds)
	if err != nil {
		s.logger.Error("获取账户信息失败",
			zap.Error(err),
			zap.String("userId", userId))
		return nil, fmt.Errorf("获取账户信息失败: %w", err)
	}

	// 使用 errgroup 进行并发同步
	g, ctx := errgroup.WithContext(ctx)
	resultChan := make(chan v1.SyncAccountResult, len(accounts))

	// 并发同步每个账户
	for _, account := range accounts {
		account := account // 创建副本用于 goroutine
		g.Go(func() error {
			syncResult := v1.SyncAccountResult{
				AccountID: account.AccountID,
			}

			// 1. 同步订阅信息
			subCount, err := s.subscriptionsService.SyncSubscriptions(ctx, userId, account.AccountID)
			if err != nil {
				syncResult.Message = fmt.Sprintf("同步订阅失败: %v", err)
				resultChan <- syncResult
				return nil // 不中断其他同步
			}
			syncResult.SubscriptionCount = subCount

			// 2. 同步虚拟机信息
			vmStats, err := s.virtualMachineService.SyncVMs(ctx, userId, account.AccountID)
			if err != nil {
				syncResult.Message = fmt.Sprintf("同步虚拟机失败: %v", err)
				resultChan <- syncResult
				return nil
			}
			syncResult.VMCount = vmStats.TotalVMs

			// 同步成功
			syncResult.Message = "同步成功"
			resultChan <- syncResult
			return nil
		})
	}

	// 等待所有同步完成
	go func() {
		err := g.Wait()
		if err != nil {
			s.logger.Error("同步账户失败", zap.Error(err))
			return
		}
		close(resultChan)
	}()

	// 收集结果
	for res := range resultChan {
		if strings.Contains(res.Message, "失败") {
			result.FailedAccounts = append(result.FailedAccounts, res)
			s.logger.Error("账户同步失败",
				zap.String("accountId", res.AccountID),
				zap.String("message", res.Message))
		} else {
			result.SuccessAccounts = append(result.SuccessAccounts, res)
			s.logger.Info("账户同步成功",
				zap.String("accountId", res.AccountID),
				zap.Int("subscriptionCount", res.SubscriptionCount),
				zap.Int("vmCount", res.VMCount))
		}
	}

	return result, nil
}
