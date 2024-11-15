package handler

import (
	"azure-vm-backend/internal/service"
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

func (h *VmRegionHandler) GetVmRegion(ctx *gin.Context) {

}
