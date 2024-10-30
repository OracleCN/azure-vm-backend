package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AccountsHandler struct {
	*Handler
	accountsService service.AccountsService
}

func NewAccountsHandler(
	handler *Handler,
	accountsService service.AccountsService,
) *AccountsHandler {
	return &AccountsHandler{
		Handler:         handler,
		accountsService: accountsService,
	}
}

func (h *AccountsHandler) GetAccounts(ctx *gin.Context) {
}

// CreateAccounts godoc
// @Summary 创建azure账户
// @Schemes
// @Description
// @Tags 账户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.GetProfileResponse
// @Router /create [post]
func (h *AccountsHandler) CreateAccounts(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}
	var req v1.CreateAccountReq
	// 获取参数 转为json
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	err := h.accountsService.CreateAccount(ctx, userId, &req)
	if err != nil {
		return
	}
}
