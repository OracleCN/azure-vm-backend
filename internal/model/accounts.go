package model

import "gorm.io/gorm"

type Accounts struct {
	gorm.Model
	AccountID          string `gorm:"column:account_id;type:varchar(32);uniqueIndex;not null" json:"account_id"`
	UserID             string `gorm:"column:user_id;type:varchar(32);index;not null" json:"user_id"`
	LoginEmail         string `gorm:"column:login_email;type:varchar(128);index;not null" json:"login_email"`
	LoginPassword      string `gorm:"column:login_password;type:varchar(128);not null" json:"login_password"`
	Remark             string `gorm:"column:remark;type:text" json:"remark"`
	AppID              string `gorm:"column:app_id;type:varchar(128);not null" json:"app_id"`
	AppPassword        string `gorm:"column:app_password;type:varchar(256);not null" json:"app_password"`
	TenantID           string `gorm:"column:tenant_id;type:varchar(128);not null" json:"tenant_id"`
	SubscriptionID     string `gorm:"column:subscription_id;type:varchar(128);not null" json:"subscription_id"`
	SubscriptionStatus string `gorm:"column:subscription_status;type:varchar(32);index;default:normal;not null" json:"subscription_status"`
}

func (m *Accounts) TableName() string {
	return "accounts"
}
