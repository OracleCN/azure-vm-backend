package v1

import (
	"azure-vm-backend/internal/model"
	"azure-vm-backend/pkg/app"
)

// CreateAccountReq 创建账户请求参数
type CreateAccountReq struct {
	LoginEmail    string `json:"loginEmail" binding:"required,email"`
	LoginPassword string `json:"loginPassword" binding:"required,min=6"`
	Remark        string `json:"remark"`
	AppID         string `json:"appId" binding:"required"`
	PassWord      string `json:"password" binding:"required"`
	Tenant        string `json:"tenant" binding:"required"`
	DisplayName   string `json:"displayName" binding:"required"`
	VmCount       int    `json:"vmCount"`
}

// CreateAccountResp 创建账户响应参数
type CreateAccountResp struct {
	AccountID  string `json:"account_id"`  // 账户ID
	LoginEmail string `json:"login_email"` // 登录邮箱
	Remark     string `json:"remark"`      // 备注
	Status     string `json:"status"`      // 状态
}

// UpdateAccountReq 在 v1 包中定义请求结构
type UpdateAccountReq struct {
	LoginEmail    string `json:"loginEmail,omitempty"`
	LoginPassword string `json:"loginPassword,omitempty"`
	Remark        string `json:"remark,omitempty"`
	AppID         string `json:"appId,omitempty"`
	PassWord      string `json:"password,omitempty"`
	Tenant        string `json:"tenant,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
}

// AccountListReq 获取账户列表的请求参数
type AccountListReq struct {
	Page     int    `json:"page" binding:"min=1"`     // 当前页码
	PageSize int    `json:"pageSize" binding:"min=1"` // 每页大小
	Search   string `json:"search"`                   // 搜索关键词(支持邮箱和备注模糊查询)
}

// AccountListResp 获取账户列表的响应
type AccountListResp struct {
	Items      []*AccountInfo `json:"items"`      // 账户列表
	Page       int            `json:"page"`       // 当前页码
	PageSize   int            `json:"pageSize"`   // 每页大小
	Total      int64          `json:"total"`      // 总记录数
	TotalPages int            `json:"totalPages"` // 总页数
}

// AccountInfo 账户信息
type AccountInfo struct {
	AccountID          string `json:"accountId"`          // 账户ID
	LoginEmail         string `json:"loginEmail"`         // 登录邮箱
	Remark             string `json:"remark"`             // 备注
	AppID              string `json:"appId"`              // Azure应用ID
	Tenant             string `json:"tenant"`             // Azure租户ID
	VmCount            int    `json:"vmCount"`            // VM数量
	DisplayName        string `json:"displayName"`        // 显示名称
	CreatedAt          string `json:"createdAt"`          // 创建时间
	UpdatedAt          string `json:"updatedAt"`          // 更新时间
	SubscriptionStatus string `json:"subscriptionStatus"` // 订阅状态
}

// ToAccountInfo 将数据库模型转换为API响应模型
func ToAccountInfo(account *model.Accounts) *AccountInfo {
	return &AccountInfo{
		AccountID:          account.AccountID,
		LoginEmail:         account.LoginEmail,
		Remark:             account.Remark,
		AppID:              account.AppID,
		Tenant:             account.Tenant,
		VmCount:            account.VmCount,
		DisplayName:        account.DisplayName,
		CreatedAt:          account.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:          account.UpdatedAt.Format("2006-01-02 15:04:05"),
		SubscriptionStatus: account.SubscriptionStatus,
	}
}

// ToAccountListResp 转换为列表响应
func ToAccountListResp(result *app.ListResult[*model.Accounts]) *AccountListResp {
	items := make([]*AccountInfo, 0, len(result.Items))
	for _, account := range result.Items {
		items = append(items, ToAccountInfo(account))
	}

	return &AccountListResp{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}
}

// SyncAccountReq 同步账户请求参数
type SyncAccountReq struct {
	AccountIds []string `json:"accountIds" binding:"required"` // 要同步的账户ID列表
}

// SyncAccountResp 同步账户响应
type SyncAccountResp struct {
	SuccessAccounts []SyncAccountResult `json:"successAccounts"` // 同步成功的账户
	FailedAccounts  []SyncAccountResult `json:"failedAccounts"`  // 同步失败的账户
}

// SyncAccountResult 单个账户的同步结果
type SyncAccountResult struct {
	AccountID         string `json:"accountId"`         // 账户ID
	Message           string `json:"message"`           // 同步信息
	SubscriptionCount int    `json:"subscriptionCount"` // 同步的订阅数量
	VMCount           int    `json:"vmCount"`           // 同步的虚拟机数量
}
