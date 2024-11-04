package handler

import (
	"azure-vm-backend/internal/service"
	"github.com/gin-gonic/gin"
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

func (h *SubscriptionsHandler) GetSubscriptions(ctx *gin.Context) {

}
