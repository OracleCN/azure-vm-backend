package app

import (
	"gorm.io/gorm"
)

// Pagination 分页参数
type Pagination struct {
	Page       int   `json:"page"`       // 当前页码
	PageSize   int   `json:"pageSize"`   // 每页大小
	Total      int64 `json:"total"`      // 总记录数
	TotalPages int   `json:"totalPages"` // 总页数
}

// QueryOption 查询选项
type QueryOption struct {
	Pagination
	SortBy    string            `json:"sortBy"`    // 排序字段
	SortOrder string            `json:"sortOrder"` // 排序方向 (asc/desc)
	Filters   map[string]string `json:"filters"`   // 过滤条件
}

// ListResult 通用列表查询结果
type ListResult[T any] struct {
	Items []T `json:"items"`
	Pagination
}

// WithPagination GORM分页查询高阶函数
func WithPagination[T any](
	db *gorm.DB,
	query *QueryOption,
	baseQuery func(*gorm.DB) *gorm.DB,
) (*ListResult[T], error) {
	var items []T
	var total int64

	// 构建基础查询
	queryDB := baseQuery(db)

	// 计算总记录数
	countQuery := queryDB.Session(&gorm.Session{}) // 创建新会话避免影响原查询
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	if total == 0 {
		return &ListResult[T]{
			Items: []T{},
			Pagination: Pagination{
				Page:       query.Page,
				PageSize:   query.PageSize,
				Total:      0,
				TotalPages: 0,
			},
		}, nil
	}

	// 验证并调整分页参数
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10 // 默认每页10条
	}

	// 计算总页数
	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	// 应用分页和排序
	offset := (query.Page - 1) * query.PageSize
	queryDB = queryDB.Offset(offset).Limit(query.PageSize)

	// 应用排序
	if query.SortBy != "" && query.SortOrder != "" {
		direction := "DESC"
		if query.SortOrder == "asc" {
			direction = "ASC"
		}
		queryDB = queryDB.Order(query.SortBy + " " + direction)
	}

	// 执行查询
	if err := queryDB.Find(&items).Error; err != nil {
		return nil, err
	}

	return &ListResult[T]{
		Items: items,
		Pagination: Pagination{
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// DefaultQueryOption 创建默认的查询选项
func DefaultQueryOption() *QueryOption {
	return &QueryOption{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
		SortOrder: "desc",
		Filters:   make(map[string]string),
	}
}

// ValidateAndFillQueryOption 验证并填充查询选项的默认值
func ValidateAndFillQueryOption(option *QueryOption) *QueryOption {
	if option == nil {
		return DefaultQueryOption()
	}

	if option.Page < 1 {
		option.Page = 1
	}
	if option.PageSize < 1 {
		option.PageSize = 10
	}
	if option.SortOrder == "" {
		option.SortOrder = "desc"
	}
	if option.Filters == nil {
		option.Filters = make(map[string]string)
	}

	return option
}
