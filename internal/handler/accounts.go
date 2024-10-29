package handler

import (
	"azure-vm-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AccountsHandler struct {
	*Handler
	accountsService service.AccountsService
}

func NewAccountsHandler(
	handler *Handler,
	accountsService service.AccountsService,
) *AccountsHandler {
	return &AccountsHandler{
		Handler:         handler,
		accountsService: accountsService,
	}
}

func (h *AccountsHandler) GetAccounts(ctx *gin.Context) {

}
