package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type VirtualMachineHandler struct {
	*Handler
	vmService service.VirtualMachineService
}

func NewVirtualMachineHandler(
	handler *Handler,
	vmService service.VirtualMachineService,
) *VirtualMachineHandler {
	return &VirtualMachineHandler{
		Handler:   handler,
		vmService: vmService,
	}
}

// GetVM godoc
// @Summary 获取单个虚拟机信息
// @Schemes
// @Description 获取指定虚拟机的详细信息
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Param vmId path string true "虚拟机ID"
// @Success 200 {object} v1.Response
// @Router /vms/{accountId}/{vmId} [get]
func (h *VirtualMachineHandler) GetVM(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取账户ID和虚拟机ID
	accountId := ctx.Param("accountId")
	vmId := ctx.Param("vmId")
	if accountId == "" || vmId == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 获取虚拟机信息
	vm, err := h.vmService.GetVM(ctx, userId, accountId, vmId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, vm)
}

// ListVMs godoc
// @Summary 查询虚拟机列表
// @Schemes
// @Description 根据查询条件获取虚拟机列表
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId query string true "账户ID"
// @Param subscriptionId query string false "订阅ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param name query string false "虚拟机名称"
// @Param resourceGroup query string false "资源组"
// @Param location query string false "地域"
// @Param status query string false "状态"
// @Param size query string false "规格"
// @Param syncStatus query string false "同步状态"
// @Success 200 {object} v1.Response
// @Router /vms [get]
func (h *VirtualMachineHandler) ListVMs(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 构建查询参数
	params := &v1.VMQueryParams{
		AccountID:      ctx.Query("accountId"),
		SubscriptionID: ctx.Query("subscriptionId"),
		Name:           ctx.Query("name"),
		ResourceGroup:  ctx.Query("resourceGroup"),
		Location:       ctx.Query("location"),
		Status:         ctx.Query("status"),
		Size:           ctx.Query("size"),
		SyncStatus:     ctx.Query("syncStatus"),
	}

	// 获取虚拟机列表
	result, err := h.vmService.ListVMs(ctx, params)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, result)
}

// ListVMsBySubscription godoc
// @Summary 获取订阅下的虚拟机列表
// @Schemes
// @Description 获取指定订阅下的所有虚拟机信息
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Param subscriptionId path string true "订阅ID"
// @Success 200 {object} v1.Response
// @Router /vms/{accountId}/{subscriptionId}/list [get]
func (h *VirtualMachineHandler) ListVMsBySubscription(ctx *gin.Context) {
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

	// 获取虚拟机列表
	vms, err := h.vmService.ListVMsBySubscription(ctx, userId, accountId, subscriptionId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, vms)
}

// SyncVMs godoc
// @Summary 同步账户下的所有虚拟机
// @Schemes
// @Description 从Azure同步指定账户下的所有虚拟机信息
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Success 200 {object} v1.Response
// @Router /vms/{accountId}/sync [post]
func (h *VirtualMachineHandler) SyncVMs(ctx *gin.Context) {
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

	// 同步虚拟机信息
	ms, err := h.vmService.SyncVMs(ctx, userId, accountId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, ms)
}

// SyncVMsBySubscription godoc
// @Summary 同步订阅下的所有虚拟机
// @Schemes
// @Description 从Azure同步指定订阅下的所有虚拟机信息
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Param subscriptionId path string true "订阅ID"
// @Success 200 {object} v1.Response
// @Router /vms/{accountId}/{subscriptionId}/sync [post]
func (h *VirtualMachineHandler) SyncVMsBySubscription(ctx *gin.Context) {
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

	// 同步虚拟机信息
	err := h.vmService.SyncVMsBySubscription(ctx, userId, accountId, subscriptionId)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// CreateVM godoc
// @Summary 创建虚拟机
// @Schemes
// @Description 在指定账户下创建新的虚拟机(暂未实现)
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Param request body v1.VMCreateParams true "创建参数"
// @Success 200 {object} v1.Response
// @Router /vms/{accountId} [post]
func (h *VirtualMachineHandler) CreateVM(ctx *gin.Context) {
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

	// 解析创建参数
	var params v1.VMCreateParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 创建虚拟机
	vm, err := h.vmService.CreateVM(ctx, userId, accountId, &params)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	v1.HandleSuccess(ctx, vm)
}

// UpdateDNSLabel godoc
// @Summary 更新虚拟机DNS标签
// @Schemes
// @Description 更新虚拟机DNS标签
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Success 200 {object} v1.Response
// @Router /vms/update/dns/{accountId}/{ID} [Post]
func (h *VirtualMachineHandler) UpdateDNSLabel(ctx *gin.Context) {
	// 获取用户id
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 获取路径参数
	accountId := ctx.Param("accountId")
	ID := ctx.Param("ID")
	if accountId == "" || ID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 解析请求体
	var req v1.UpdateDNSLabelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 参数验证
	if req.DNSLabel == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 调用服务层更新DNS标签
	err := h.vmService.UpdateDNSLabel(ctx, userId, accountId, ID, req.DNSLabel)
	if err != nil {
		switch {
		case errors.Is(err, v1.ErrUnauthorized):
			v1.HandleError(ctx, http.StatusUnauthorized, err, nil)
		case errors.Is(err, v1.ErrAccountError):
			v1.HandleError(ctx, http.StatusBadRequest, err, nil)
		case errors.Is(err, v1.ErrorAzureNotFound):
			v1.HandleError(ctx, http.StatusNotFound, err, nil)
		default:
			v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		}
		return
	}

	// 成功响应
	v1.HandleSuccess(ctx, nil)
}

// OperateVM godoc
// @Summary 执行虚拟机操作
// @Schemes
// @Description 对虚拟机执行指定操作（启动/停止/重启/删除）
// @Tags 虚拟机模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param accountId path string true "账户ID"
// @Param vmId path string true "虚拟机ID"
// @Param request body v1.VMOperationRequest true "操作请求"
// @Success 200 {object} v1.Response
// @Router /vms/{accountId}/{vmId}/operate [post]
func (h *VirtualMachineHandler) OperateVM(ctx *gin.Context) {
	// 1. 获取用户ID
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 2. 获取路径参数
	accountId := ctx.Param("accountId")
	id := ctx.Param("id")
	if accountId == "" || id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 3. 解析请求体
	var req v1.VMOperationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 4. 执行操作
	err := h.vmService.OperateVM(ctx, userId, accountId, id, req.Operation, req.Force)
	if err != nil {
		switch {
		case errors.Is(err, v1.ErrUnauthorized):
			v1.HandleError(ctx, http.StatusUnauthorized, err, nil)
		case errors.Is(err, v1.ErrAccountError), errors.Is(err, v1.ErrBadRequest):
			v1.HandleError(ctx, http.StatusBadRequest, err, nil)
		case errors.Is(err, v1.ErrorAzureNotFound):
			v1.HandleError(ctx, http.StatusNotFound, err, nil)
		default:
			v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		}
		return
	}

	// 5. 返回成功
	v1.HandleSuccess(ctx, nil)
}
