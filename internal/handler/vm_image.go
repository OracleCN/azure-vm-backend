package handler

import (
	"azure-vm-backend/internal/service"
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

func (h *VmImageHandler) GetVmImage(ctx *gin.Context) {

}
