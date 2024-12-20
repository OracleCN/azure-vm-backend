package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"azure-vm-backend/pkg/app"
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

// ListAccounts godoc
// @Summary 获取账户列表
// @Schemes
// @Description 获取当前用户的所有Azure账户
// @Tags 账户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.Response
// @Router /accounts/list [post]
func (h *AccountsHandler) ListAccounts(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	var req v1.AccountListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 创建查询选项
	option := &app.QueryOption{
		Pagination: app.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Filters: map[string]string{
			"search": req.Search,
		},
	}

	// 验证并填充默认值
	option = app.ValidateAndFillQueryOption(option)

	// 调用service层获取账户列表
	result, err := h.accountsService.GetAccountList(ctx, userId, option)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	// 转换并返回响应
	resp := v1.ToAccountListResp(result)
	v1.HandleSuccess(ctx, resp)
}

// GetAccount godoc
// @Summary 获取单个账户信息
// @Schemes
// @Description 获取指定Azure账户的详细信息
// @Tags 账户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "账户ID"
// @Success 200 {object} v1.Response
// @Router /accounts/{id} [get]
func (h *AccountsHandler) GetAccount(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID
	accountId := ctx.Param("id")
	if accountId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 获取账户信息
	account, err := h.accountsService.GetAccount(ctx, userId, accountId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	// 返回成功响应
	v1.HandleSuccess(ctx, account)
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
	accountId, err := h.accountsService.CreateAccount(ctx, userId, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, map[string]string{
		"accountId": accountId,
	})
}

// UpdateAccount godoc
// @Summary 更新Azure账户
// @Schemes
// @Description 更新指定的Azure账户信息
// @Tags 账户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "账户ID"
// @Param request body v1.UpdateAccountRequest true "更新账户请求参数"
// @Success 200 {object} v1.AzureAccount
// @Router /accounts/{id} [post]
func (h *AccountsHandler) UpdateAccount(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID
	accountId := ctx.Param("id")
	if accountId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	var req v1.UpdateAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err)
		return
	}

	// 更新账户
	err := h.accountsService.UpdateAccount(ctx, userId, accountId, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	// 返回成功响应
	v1.HandleSuccess(ctx, nil)
}

// DeleteAccounts godoc
// @Summary 批量删除Azure账户
// @Schemes
// @Description 批量删除指定的Azure账户
// @Tags 账户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param ids query string true "账户ID列表,多个ID用逗号分隔"
// @Success 200 {object} v1.Response
// @Router /accounts/delete [delete]
func (h *AccountsHandler) DeleteAccounts(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 直接绑定请求体中的数组
	var accountIds []string
	if err := ctx.ShouldBindJSON(&accountIds); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	if len(accountIds) == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 批量删除账户
	err := h.accountsService.DeleteAccount(ctx, userId, accountIds)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	// 返回成功响应
	v1.HandleSuccess(ctx, gin.H{
		"deletedCount": "ok",
	})
}

// SyncAccounts godoc
// @Summary 同步Azure账户
// @Schemes
// @Description 同步Azure账户
// @Tags 账户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param ids query string true "[]"
// @Success 200 {object} v1.Response
// @Router /accounts/delete [post]
func (h *AccountsHandler) SyncAccounts(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	var req v1.SyncAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 调用service层进行同步
	result, err := h.accountsService.SyncAccounts(ctx, userId, req.AccountIds)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, result)
}
