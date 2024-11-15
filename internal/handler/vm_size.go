package handler

import (
	"azure-vm-backend/internal/service"
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

func (h *VmSizeHandler) GetVmSize(ctx *gin.Context) {

}
