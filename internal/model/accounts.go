package model

import "gorm.io/gorm"

type Accounts struct {
	gorm.Model
	AccountID          string `gorm:"column:account_id;type:varchar(32);uniqueIndex;not null" json:"accountId"`
	UserID             string `gorm:"column:user_id;type:varchar(32);index;not null" json:"userId"`
	LoginEmail         string `gorm:"column:login_email;type:varchar(128);index;not null" json:"loginEmail"`
	LoginPassword      string `gorm:"column:login_password;type:varchar(128);not null" json:"loginPassword"`
	Remark             string `gorm:"column:remark;type:text" json:"remark"`
	AppID              string `gorm:"column:app_id;type:varchar(128);not null" json:"appId"`
	PassWord           string `gorm:"column:password;type:varchar(256);not null" json:"password"`
	Tenant             string `gorm:"column:tenant;type:varchar(128);not null" json:"tenant"`
	DisplayName        string `gorm:"column:display_name;type:varchar(128);not null" json:"displayName"`
	SubscriptionStatus string `gorm:"column:subscription_status;type:varchar(32);index;default:normal;not null" json:"subscription_status"`
}

func (m *Accounts) TableName() string {
	return "accounts"
}
