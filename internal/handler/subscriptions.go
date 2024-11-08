package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SubscriptionsHandler struct {
	*Handler
	subscriptionsService service.SubscriptionsService
}

func NewSubscriptionsHandler(
	handler *Handler,
	subscriptionsService service.SubscriptionsService,
) *SubscriptionsHandler {
	return &SubscriptionsHandler{
		Handler:              handler,
		subscriptionsService: subscriptionsService,
	}
}

// GetSubscriptions godoc
// @Summary 获取账户订阅列表
// @Schemes
// @Description 获取指定Azure账户的所有订阅信息
// @Tags 订阅模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Success 200 {object} v1.Response
// @Router /subscriptions/{accountId} [post]
func (h *SubscriptionsHandler) GetSubscriptions(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID
	accountId := ctx.Param("accountId")
	if accountId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 获取订阅列表
	subscriptions, err := h.subscriptionsService.GetSubscriptions(ctx, userId, accountId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, subscriptions)
}

// GetSubscription godoc
// @Summary 获取单个订阅信息
// @Schemes
// @Description 获取指定订阅的详细信息
// @Tags 订阅模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Param subscriptionId path string true "订阅ID"
// @Success 200 {object} v1.Response
// @Router /subscriptions/{accountId}/{subscriptionId} [get]
func (h *SubscriptionsHandler) GetSubscription(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID和订阅ID
	accountId := ctx.Param("accountId")
	subscriptionId := ctx.Param("subscriptionId")
	if accountId == "" || subscriptionId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 获取订阅信息
	subscription, err := h.subscriptionsService.GetSubscription(ctx, userId, accountId, subscriptionId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, subscription)
}

// SyncSubscriptions godoc
// @Summary 同步账户订阅信息
// @Schemes
// @Description 从Azure同步指定账户的最新订阅信息
// @Tags 订阅模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Success 200 {object} v1.Response
// @Router /subscriptions/{accountId}/sync [post]
func (h *SubscriptionsHandler) SyncSubscriptions(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID
	accountId := ctx.Param("accountId")
	if accountId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 同步订阅信息
	count, err := h.subscriptionsService.SyncSubscriptions(ctx, userId, accountId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, map[string]int{
		"count": count,
	})
}

// DeleteSubscriptions godoc
// @Summary 删除账户订阅信息
// @Schemes
// @Description 删除指定账户的所有订阅信息
// @Tags 订阅模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Success 200 {object} v1.Response
// @Router /subscriptions/{accountId} [delete]
func (h *SubscriptionsHandler) DeleteSubscriptions(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID
	accountId := ctx.Param("accountId")
	if accountId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 删除订阅信息
	err := h.subscriptionsService.DeleteSubscriptions(ctx, userId, accountId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}
