package service

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/pkg/azure"
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type AccountsService interface {
	// CreateAccount 创建账户
	CreateAccount(ctx context.Context, userId string, req *v1.CreateAccountReq) error
}

func NewAccountsService(
	service *Service,
	accountsRepo repository.AccountsRepository,
) AccountsService {
	return &accountsService{
		Service:      service,
		accountsRepo: accountsRepo,
	}
}

type accountsService struct {
	*Service
	accountsRepo repository.AccountsRepository
}

func (s *accountsService) CreateAccount(ctx context.Context, userId string, req *v1.CreateAccountReq) error {
	// 判断 邮箱是否重复出现在 账号表内
	existingAccount, err := s.accountsRepo.GetAccountByEmail(ctx, req.LoginEmail)
	if err != nil {
		s.logger.Error("未检查电子邮件是否存在",
			zap.Error(err),
			zap.String("email", req.LoginEmail),
		)
		return v1.ErrAccountError
	}
	if existingAccount != nil {
		s.logger.Warn("已使用的电子邮件",
			zap.String("email", req.LoginEmail),
			zap.String("existing_user_id", existingAccount.UserID),
		)
		return v1.ErrAccountEmailDuplicate
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
		return fmt.Errorf("azure验证失败: %s", result.Message)
	}
	// 获取订阅类型

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
		SubscriptionStatus: "normal",
	}

	if err := s.accountsRepo.Create(ctx, account); err != nil {
		s.logger.Error("failed to create account",
			zap.Error(err),
			zap.String("user_id", userId),
			zap.String("login_email", req.LoginEmail),
		)
		return fmt.Errorf("创建账户失败: %w", err)
	}

	s.logger.Info("帐户已成功创建",
		zap.String("account_id", account.AccountID),
		zap.String("user_id", userId),
		zap.String("login_email", req.LoginEmail),
	)

	return nil
}
