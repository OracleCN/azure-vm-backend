package v1

// CreateAccountReq 创建账户请求参数
type CreateAccountReq struct {
	LoginEmail    string `json:"loginEmail" binding:"required,email"`    // 登录邮箱
	LoginPassword string `json:"loginPassword" binding:"required,min=6"` // 登录密码
	Remark        string `json:"remark"`                                 // 备注
	AppID         string `json:"appId" binding:"required"`               // 应用ID
	PassWord      string `json:"password" binding:"required"`            // 应用密码
	Tenant        string `json:"tenant" binding:"required"`              // 租户ID
	DisplayName   string `json:"displayName" binding:"required"`         // 订阅ID
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
