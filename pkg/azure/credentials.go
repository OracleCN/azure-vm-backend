package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// AzureCredential Azure认证信息
type AzureCredential struct {
	TenantID     string // Azure租户ID
	ClientID     string // 应用程序(客户端)ID
	ClientSecret string // 客户端密钥
	DisplayName  string // 显示名称（可选）
}

// GetCredential 获取Azure认证对象
func (c *AzureCredential) GetCredential() (*azidentity.ClientSecretCredential, error) {
	if c.TenantID == "" || c.ClientID == "" || c.ClientSecret == "" {
		return nil, fmt.Errorf("缺少必要的认证信息")
	}

	credential, err := azidentity.NewClientSecretCredential(
		c.TenantID,
		c.ClientID,
		c.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("创建Azure认证对象失败: %w", err)
	}

	return credential, nil
}

// Validate 验证认证信息是否完整
func (c *AzureCredential) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if c.ClientID == "" {
		return fmt.Errorf("客户端ID不能为空")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("客户端密钥不能为空")
	}
	return nil
}
