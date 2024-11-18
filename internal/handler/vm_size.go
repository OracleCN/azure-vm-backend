package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type VmSizeHandler struct {
	*Handler
	vmSizeService service.VmSizeService
}

func NewVmSizeHandler(
	handler *Handler,
	vmSizeService service.VmSizeService,
) *VmSizeHandler {
	return &VmSizeHandler{
		Handler:       handler,
		vmSizeService: vmSizeService,
	}
}

// ListVmSizes 获取规格列表
func (h *VmSizeHandler) ListVmSizes(ctx *gin.Context) {
	var req v1.ListVmSizesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 如果没有指定location，使用默认值
	if req.Location == "" {
		req.Location = "eastasia"
	}

	sizes, err := h.vmSizeService.ListVmSizes(ctx.Request.Context(), req.Location)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	response := v1.ToListVmSizesResponse(sizes)
	v1.HandleSuccess(ctx, response)
}

// SyncVmSizes 同步规格信息
func (h *VmSizeHandler) SyncVmSizes(ctx *gin.Context) {
	// 获取用户ID
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	var req v1.SyncVmSizesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	err := h.vmSizeService.SyncVmSizes(
		ctx.Request.Context(),
		userId,
		req.AccountID,
		req.SubscriptionID,
		req.Location,
	)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	response := &v1.SyncVmSizesResponse{
		Message:  "同步成功",
		SyncTime: time.Now().Format(time.RFC3339),
		Location: req.Location,
	}

	v1.HandleSuccess(ctx, response)
}
