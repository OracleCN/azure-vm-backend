package v1

// ListSubscriptionsRequest 定义请求体结构
type ListSubscriptionsRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
}

// SyncRequest represents the request body for bulk subscription sync
type SyncRequest struct {
	AccountIDs []string `json:"accountIds" binding:"required,min=1"`
}

// SyncResult represents the sync result for a single account
type SyncResult struct {
	AccountID string `json:"accountId"`
	Count     int    `json:"count"`
	Error     string `json:"error,omitempty"`
}

// BulkSyncResponse represents the response for bulk subscription sync
type BulkSyncResponse struct {
	Results []SyncResult `json:"results"`
	Summary struct {
		TotalAccounts int `json:"totalAccounts"`
		SuccessCount  int `json:"successCount"`
		FailureCount  int `json:"failureCount"`
		TotalSynced   int `json:"totalSynced"`
	} `json:"summary"`
}
