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
	AccountID            string     `gorm:"column:account_id;type:varchar(32);not null;uniqueIndex:idx_account_subscription"`
	SubscriptionID       string     `gorm:"column:subscription_id;type:varchar(32);not null;uniqueIndex:idx_account_subscription"`
	DisplayName          string     `gorm:"column:display_name;type:varchar(128);not null"`
	State                string     `gorm:"column:state;type:varchar(32);not null"`
	SubscriptionPolicies string     `gorm:"column:subscription_policies;type:text"`
	AuthorizationSource  string     `gorm:"column:authorization_source;type:varchar(32)"`
	SubscriptionType     string     `gorm:"column:subscription_type;type:varchar(32)"`
	SpendingLimit        string     `gorm:"column:spending_limit;type:varchar(32)"`
	StartDate            *time.Time `gorm:"column:start_date"`
	EndDate              *time.Time `gorm:"column:end_date"`
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
