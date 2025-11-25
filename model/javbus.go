package model

// ==========================================
// 枚举与类型定义 (Enums & Type Aliases)
// ==========================================

type MovieType string

const (
	MovieTypeNormal     MovieType = "normal"
	MovieTypeUncensored MovieType = "uncensored"
)

type MagnetType string

const (
	MagnetTypeAll   MagnetType = "all"
	MagnetTypeExist MagnetType = "exist"
)

type FilterType string

const (
	FilterTypeStar     FilterType = "star"
	FilterTypeGenre    FilterType = "genre"
	FilterTypeDirector FilterType = "director"
	FilterTypeStudio   FilterType = "studio"
	FilterTypeLabel    FilterType = "label"
	FilterTypeSeries   FilterType = "series"
)

type SortBy string

const (
	SortByDate SortBy = "date"
	SortBySize SortBy = "size"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// ==========================================
// 基础数据结构 (Basic Structures)
// ==========================================

type Property struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ==========================================
// 电影相关结构 (Movie Structures)
// ==========================================

type Movie struct {
	Date  *string  `json:"date"` // nullable
	Title string   `json:"title"`
	ID    string   `json:"id"`
	Img   *string  `json:"img"` // nullable
	Tags  []string `json:"tags"`
}

type SimilarMovie struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Img   *string `json:"img"` // nullable
}

type Sample struct {
	Alt       *string `json:"alt"` // nullable
	ID        string  `json:"id"`
	Thumbnail string  `json:"thumbnail"`
	Src       *string `json:"src"` // nullable
}

type MovieDetail struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Img           *string        `json:"img"`         // nullable
	Date          *string        `json:"date"`        // nullable
	VideoLength   *int           `json:"videoLength"` // nullable
	Director      *Property      `json:"director"`    // nullable
	Producer      *Property      `json:"producer"`    // nullable
	Publisher     *Property      `json:"publisher"`   // nullable
	Series        *Property      `json:"series"`      // nullable
	Genres        []Property     `json:"genres"`
	Stars         []Property     `json:"stars"`
	ImageSize     *ImageSize     `json:"imageSize"` // nullable
	Samples       []Sample       `json:"samples"`
	SimilarMovies []SimilarMovie `json:"similarMovies"`
	GID           *string        `json:"gid"` // nullable
	UC            *string        `json:"uc"`  // nullable
}

// ==========================================
// 磁力链接结构 (Magnet Structure)
// ==========================================

type Magnet struct {
	ID          string  `json:"id"`
	Link        string  `json:"link"`
	IsHD        bool    `json:"isHD"`
	Title       string  `json:"title"`
	Size        *string `json:"size"`       // nullable
	NumberSize  *int64  `json:"numberSize"` // nullable (Go int64 is safer for file sizes)
	ShareDate   *string `json:"shareDate"`  // nullable
	HasSubtitle bool    `json:"hasSubtitle"`
}

// ==========================================
// 演员信息结构 (Star Info Structure)
// ==========================================

type StarInfo struct {
	Avatar   *string `json:"avatar"` // nullable
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Birthday *string `json:"birthday"` // nullable
	Age      *int    `json:"age"`      // nullable
	Height   *int    `json:"height"`   // nullable

	// 胸围
	Bust *string `json:"bust"` // nullable

	// 腰围
	Waistline *string `json:"waistline"` // nullable

	// 臀围
	Hipline *string `json:"hipline"` // nullable

	Birthplace *string `json:"birthplace"` // nullable
	Hobby      *string `json:"hobby"`      // nullable
}

// ==========================================
// 分页与响应结构 (Pagination & Responses)
// ==========================================

type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	HasNextPage bool  `json:"hasNextPage"`
	NextPage    *int  `json:"nextPage"` // nullable
	Pages       []int `json:"pages"`
}

type MoviesPage struct {
	Movies     []Movie    `json:"movies"`
	Pagination Pagination `json:"pagination"`
}

// SearchMoviesPage 组合了 MoviesPage (类似于 TypeScript 的 extend)
type SearchMoviesPage struct {
	MoviesPage        // Embedded struct
	Keyword    string `json:"keyword"`
}

// ==========================================
// 请求参数结构 (Request Query)
// ==========================================

type GetMoviesQuery struct {
	Page        string     `form:"page" json:"page"` // query param usually comes as string
	Type        MovieType  `form:"type" json:"type"`
	Magnet      MagnetType `form:"magnet" json:"magnet"`
	FilterType  FilterType `form:"filterType" json:"filterType"`
	FilterValue string     `form:"filterValue" json:"filterValue"`
}
