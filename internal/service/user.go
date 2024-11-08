package service

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/model"
	"azure-vm-backend/internal/repository"
	"context"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserService interface {
	Register(ctx context.Context, req *v1.RegisterRequest) error
	Login(ctx context.Context, req *v1.LoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error)
	UpdateProfile(ctx context.Context, userId string, req *v1.UpdateProfileRequest) error
}
type UserUpdater func(user *model.User) error

func NewUserService(
	service *Service,
	userRepo repository.UserRepository,
) UserService {
	return &userService{
		userRepo: userRepo,
		Service:  service,
	}
}

type userService struct {
	userRepo repository.UserRepository
	*Service
}

func (s *userService) Register(ctx context.Context, req *v1.RegisterRequest) error {
	// check username
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return v1.ErrInternalServerError
	}
	if user != nil {
		return v1.ErrUserAlreadyExist
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Generate user ID
	userId, err := s.sid.GenString()
	if err != nil {
		return err
	}
	user = &model.User{
		UserId:   userId,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	// Transaction demo
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		// Create a user
		if err = s.userRepo.Create(ctx, user); err != nil {
			return err
		}
		// TODO: other repo
		return nil
	})
	return err
}

func (s *userService) Login(ctx context.Context, req *v1.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return "", v1.ErrUnauthorized
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", err
	}
	token, err := s.jwt.GenToken(user.UserId, time.Now().Add(time.Hour*24*1))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error) {
	user, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.GetProfileResponseData{
		UserId:   user.UserId,
		Nickname: user.Nickname,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Roles:    []string{"admin"},
	}, nil
}
func (s *userService) UpdateProfile(ctx context.Context, userId string, req *v1.UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return err
	}

	updaters := []UserUpdater{
		s.updateBasicInfo(req),
		s.updatePassword(req),
	}

	for _, update := range updaters {
		if err := update(user); err != nil {
			return err
		}
	}

	return s.userRepo.Update(ctx, user)
}

// 更新基本信息
func (s *userService) updateBasicInfo(req *v1.UpdateProfileRequest) UserUpdater {
	return func(user *model.User) error {
		if req.Email != "" {
			user.Email = req.Email
		}
		if req.Nickname != "" {
			user.Nickname = req.Nickname
		}
		if req.Avatar != "" {
			user.Avatar = req.Avatar
		}
		return nil
	}
}

// 更新密码
func (s *userService) updatePassword(req *v1.UpdateProfileRequest) UserUpdater {
	return func(user *model.User) error {
		// 如果没有新密码，跳过密码更新
		if req.ConfirmPassword == "" {
			return nil
		}

		// 验证旧密码存在
		if req.OldPassword == "" {
			return v1.ErrPasswordError
		}

		// 验证旧密码正确性
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			return v1.ErrPasswordError
		}

		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.ConfirmPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user.Password = string(hashedPassword)
		return nil
	}
}
