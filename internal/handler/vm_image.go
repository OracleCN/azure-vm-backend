package handler

import (
	v1 "azure-vm-backend/api/v1"
	"azure-vm-backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type VmImageHandler struct {
	*Handler
	vmImageService service.VmImageService
}

func NewVmImageHandler(
	handler *Handler,
	vmImageService service.VmImageService,
) *VmImageHandler {
	return &VmImageHandler{
		Handler:        handler,
		vmImageService: vmImageService,
	}
}

// GetVmImage 获取指定ID的镜像信息
func (h *VmImageHandler) GetVmImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	image, err := h.vmImageService.GetVmImage(ctx.Request.Context(), uint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	if image == nil {
		v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, nil)
		return
	}

	response := &v1.GetImageResponse{
		ImageInfo: *v1.ToImageInfo(image),
	}
	v1.HandleSuccess(ctx, response)
}

// ListVmImages 获取所有可用的镜像列表
func (h *VmImageHandler) ListVmImages(ctx *gin.Context) {
	images, err := h.vmImageService.ListVmImages(ctx.Request.Context())
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	// 转换为响应格式
	var imageInfos []*v1.ImageInfo
	for _, image := range images {
		imageInfos = append(imageInfos, v1.ToImageInfo(image))
	}

	response := &v1.ListImagesResponse{
		Total:  int64(len(images)),
		Images: imageInfos,
	}

	v1.HandleSuccess(ctx, response)
}

// SyncVmImages godoc
// @Summary 同步Azure镜像信息
// @Description 从Azure同步指定区域的镜像信息
// @Tags 镜像管理
// @Accept json
// @Produce json
// @Param request body v1.SyncImagesRequest true "同步请求参数"
// @Success 200 {object} v1.Response{data=v1.SyncImagesResponse}
// @Router /api/v1/images/sync [post]
func (h *VmImageHandler) SyncVmImages(ctx *gin.Context) {
	// 获取用户ID
	userId := GetUserIdFromCtx(ctx)
	if userId == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
		return
	}

	// 解析请求体
	var req v1.SyncImagesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}

	// 调用服务层同步镜像
	if err := h.vmImageService.SyncVmImages(ctx.Request.Context(), userId, req.AccountID, req.SubscriptionID, req.Location); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	// 构造响应
	response := &v1.SyncImagesResponse{
		Message:  "同步成功",
		SyncTime: time.Now().Format(time.RFC3339),
		Location: req.Location,
	}

	v1.HandleSuccess(ctx, response)
}
