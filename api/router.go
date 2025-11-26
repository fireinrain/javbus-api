package api

import (
	"net/http"
	"strings"

	"github.com/fireinrain/javbus-api/config"
	"github.com/fireinrain/javbus-api/scraper"
	"github.com/gin-gonic/gin"

	"github.com/fireinrain/javbus-api/model" // 引用你的 model 包
)

var JavbusScraper *scraper.JavbusScraper

// RegisterRoutes 注册所有路由
// 对应 export default router
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	//初始化scraper
	javbusScraper := scraper.NewJavbusScraper(cfg)
	JavbusScraper = javbusScraper

	//是否可以访问javbus
	r.GET("/accessJavbus", GetAccessJavbus)
	movies := r.Group("/movies")
	{
		movies.GET("/", GetMovies)
		movies.GET("/search", SearchMovies)
		movies.GET("/:id", GetMovieDetail)
	}

	// 挂载 /stars 路由组
	stars := r.Group("/stars")
	{
		stars.GET("/:id", GetStarInfo)
	}

	// 挂载 /magnets 路由组
	magnets := r.Group("/magnets")
	{
		magnets.GET("/:movieId", GetMovieMagnets)
	}

}

func GetAccessJavbus(c *gin.Context) {
	resp, _ := JavbusScraper.GetAccessJavbus()
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Handlers (对应原来的 router.get 回调)
// ==========================================

// GetMovies 获取电影列表
// GET /movies
func GetMovies(c *gin.Context) {
	var query model.GetMoviesQuery
	// 对应 validate(moviesPageValidator)
	// Gin 会自动根据 struct tag 验证参数
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "messages": []string{"Invalid query parameters"}})
		return
	}

	// 调用 scraper
	resp, err := JavbusScraper.GetMoviesByPage(&query)
	if err != nil {
		c.Error(err) // 记录错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SearchMovies 搜索电影
// GET /movies/search
func SearchMovies(c *gin.Context) {
	// 定义搜索专用的 Query 结构体 (继承基础查询)
	type SearchQuery struct {
		model.GetMoviesQuery
		Keyword string `form:"keyword" binding:"required"` // 必填
	}

	var query SearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "messages": []string{"Keyword is required"}})
		return
	}

	// 调用 scraper
	resp, err := JavbusScraper.GetMoviesByKeywordAndPage(strings.TrimSpace(query.Keyword), &query.GetMoviesQuery)

	if err != nil {
		// === 复刻 Node.js 的特殊逻辑 ===
		// if (e.message.includes('404')) { 返回空列表 }
		if strings.Contains(err.Error(), "404") {
			// 构造一个空的 SearchMoviesPage 响应
			emptyResp := model.SearchMoviesPage{
				MoviesPage: model.MoviesPage{
					Movies: []model.Movie{},
					Pagination: model.Pagination{
						CurrentPage: 1, // 这里简化处理，也可以尝试解析 page 参数
						HasNextPage: false,
						Pages:       []int{},
					},
				},
				Keyword: query.Keyword,
			}
			c.JSON(http.StatusOK, emptyResp)
			return
		}

		// 其他错误抛出
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMovieDetail 获取电影详情
// GET /movies/:id
func GetMovieDetail(c *gin.Context) {
	movieId := c.Param("id")

	movie, err := JavbusScraper.GetMovieDetail(movieId)
	if err != nil {
		// 复刻 404 处理逻辑
		if strings.Contains(err.Error(), "404") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, movie)
}

// GetStarInfo 获取演员信息
// GET /stars/:id
func GetStarInfo(c *gin.Context) {
	starId := c.Param("id")

	// 获取 type 参数 (normal/uncensored)
	movieType := c.Query("type")

	starInfo, err := JavbusScraper.GetStarInfo(starId, movieType)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, starInfo)
}

// GetMovieMagnets 获取磁力链接
// GET /magnets/:movieId
func GetMovieMagnets(c *gin.Context) {
	movieId := c.Param("movieId")

	// 定义请求参数结构体
	// 对应 Node: const { gid, uc, sortBy, sortOrder } = req.query;
	var query struct {
		GID       string `form:"gid"`
		UC        string `form:"uc"`
		SortBy    string `form:"sortBy"`
		SortOrder string `form:"sortOrder"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}

	magnets, err := JavbusScraper.GetMovieMagnets(movieId, query.GID, query.UC, query.SortBy, query.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, magnets)
}
