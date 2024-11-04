package model

import (
	"azure-vm-backend/pkg/azure"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

// Subscriptions Subscription 订阅信息模型
type Subscriptions struct {
	gorm.Model
	AccountID            string     `gorm:"column:account_id;type:varchar(32);not null" json:"accountId"`
	SubscriptionID       string     `gorm:"column:subscription_id;type:varchar(36);not null" json:"subscriptionId"`
	DisplayName          string     `gorm:"column:display_name;type:varchar(128);not null" json:"displayName"`
	State                string     `gorm:"column:state;type:varchar(32);not null" json:"state"`
	SubscriptionPolicies string     `gorm:"column:subscription_policies;type:text" json:"subscriptionPolicies"`
	AuthorizationSource  string     `gorm:"column:authorization_source;type:varchar(32)" json:"authorizationSource"`
	SubscriptionType     string     `gorm:"column:subscription_type;type:varchar(32)" json:"subscriptionType"`
	SpendingLimit        string     `gorm:"column:spending_limit;type:varchar(32)" json:"spendingLimit"`
	StartDate            *time.Time `gorm:"column:start_date" json:"startDate"`
	EndDate              *time.Time `gorm:"column:end_date" json:"endDate"`
}

func (s *Subscriptions) TableName() string {
	return "subscriptions"
}

// SetSubscriptionPolicies 设置订阅策略JSON
func (s *Subscriptions) SetSubscriptionPolicies(policies map[string]interface{}) error {
	if policies == nil {
		s.SubscriptionPolicies = ""
		return nil
	}
	data, err := json.Marshal(policies)
	if err != nil {
		return err
	}
	s.SubscriptionPolicies = string(data)
	return nil
}

// GetSubscriptionPolicies 获取订阅策略
func (s *Subscriptions) GetSubscriptionPolicies() (map[string]interface{}, error) {
	policies := make(map[string]interface{})
	if s.SubscriptionPolicies == "" {
		return policies, nil
	}
	err := json.Unmarshal([]byte(s.SubscriptionPolicies), &policies)
	return policies, err
}

// FromAzureSubscription 从Azure SDK的订阅详情转换为本地模型
func (s *Subscriptions) FromAzureSubscription(accountID string, azureSub *azure.SubscriptionDetail) error {
	s.AccountID = accountID
	s.SubscriptionID = azureSub.SubscriptionID
	s.DisplayName = azureSub.DisplayName
	s.State = azureSub.State
	s.AuthorizationSource = azureSub.AuthorizationSource
	s.SubscriptionType = azureSub.SubscriptionType
	s.SpendingLimit = azureSub.SpendingLimit
	s.StartDate = azureSub.StartDate
	s.EndDate = azureSub.EndDate

	// 设置订阅策略
	if err := s.SetSubscriptionPolicies(azureSub.SubscriptionPolicies); err != nil {
		return err
	}

	return nil
}
