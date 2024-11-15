package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VmRegionHandler struct {
	*Handler
	vmRegionService service.VmRegionService
}

func NewVmRegionHandler(
	handler *Handler,
	vmRegionService service.VmRegionService,
) *VmRegionHandler {
	return &VmRegionHandler{
		Handler:         handler,
		vmRegionService: vmRegionService,
	}
}

// GetVmRegion godoc
// @Summary 获取单个区域信息
// @Description 获取指定ID的区域详细信息
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param id path int true "区域ID"
// @Success 200 {object} v1.Response{data=v1.VmRegionResp}
// @Router /vm/regions/{id} [get]
func (h *VmRegionHandler) GetVmRegion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrInvalidParams, fmt.Errorf("无效的ID格式: %s", idStr))
		return
	}

	region, err := h.vmRegionService.GetVmRegion(ctx, id)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err)
		return
	}

	v1.HandleSuccess(ctx, v1.ToVmRegionResp(region))
}

// ListVmRegions godoc
// @Summary 获取区域列表
// @Description 获取所有区域信息，支持启用状态筛选
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param enabled query bool false "是否只返回启用的区域"
// @Success 200 {object} v1.Response{data=v1.ListVmRegionResp}
// @Router /vm/regions [get]
func (h *VmRegionHandler) ListVmRegions(ctx *gin.Context) {
	var req v1.ListVmRegionReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrInvalidParams, err)
		return
	}

	regions, err := h.vmRegionService.ListVmRegions(ctx, req.Enabled)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err)
		return
	}

	v1.HandleSuccess(ctx, v1.ToVmRegionListResp(regions))
}
