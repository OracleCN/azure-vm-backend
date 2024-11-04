package handler

import (
	"azure-vm-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type VirtualMachineHandler struct {
	*Handler
	virtualMachineService service.VirtualMachineService
}

func NewVirtualMachineHandler(
	handler *Handler,
	virtualMachineService service.VirtualMachineService,
) *VirtualMachineHandler {
	return &VirtualMachineHandler{
		Handler:               handler,
		virtualMachineService: virtualMachineService,
	}
}

func (h *VirtualMachineHandler) GetVirtualMachine(ctx *gin.Context) {

}
