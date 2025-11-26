package model

// BaseMoviesQuery 包含分页和通用过滤参数
// 对应 baseMoviesPageValidator
type BaseMoviesQuery struct {
	// page: 对应 .isInt({ min: 1 })
	// binding:"omitempty,min=1" 表示可选，如果存在则必须 >= 1
	Page int `form:"page" binding:"omitempty,min=1"`

	// magnet: 对应 .isIn(['all', 'exist'])
	Magnet string `form:"magnet" binding:"omitempty,oneof=all exist"`

	// type: 对应 typeValidator
	Type string `form:"type" binding:"omitempty,oneof=normal uncensored"`
}

// MoviesPageQuery 首页/分类页查询
// 对应 moviesPageValidator
type MoviesPageQuery struct {
	BaseMoviesQuery

	// filterType: 对应 .isIn([...])
	// required_with=FilterValue 实现原本的 .custom(...) 逻辑：如果有 Value 则必须有 Type
	FilterType string `form:"filterType" binding:"omitempty,oneof=star genre director studio label series,required_with=FilterValue"`

	// filterValue:
	// required_with=FilterType 实现逻辑：如果有 Type 则必须有 Value
	FilterValue string `form:"filterValue" binding:"omitempty,required_with=FilterType"`
}

// SearchMoviesQuery 搜索页查询
// 对应 searchMoviesPageValidator
type SearchMoviesQuery struct {
	BaseMoviesQuery

	// keyword: 对应 .notEmpty()
	// required 表示必填
	Keyword string `form:"keyword" binding:"required"`
}

// MagnetsQuery 磁力链接查询
// 对应 magnetsValidator
type MagnetsQuery struct {
	// gid, uc: 必填
	GID string `form:"gid" binding:"required"`
	UC  string `form:"uc" binding:"required"`

	// sortBy: 对应 .isIn(['date', 'size'])
	// required_with=SortOrder: 如果有 SortOrder，则 SortBy 必填
	SortBy string `form:"sortBy" binding:"omitempty,oneof=date size,required_with=SortOrder"`

	// sortOrder: 对应 .isIn(['asc', 'desc'])
	// required_with=SortBy: 如果有 SortBy，则 SortOrder 必填
	SortOrder string `form:"sortOrder" binding:"omitempty,oneof=asc desc,required_with=SortBy"`
}
