package api

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/fireinrain/javbus-api/assets"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"

	"github.com/fireinrain/javbus-api/config" // 替换为实际路径
	//"your-project/internal/handler"           // 假设原本的 router.js 逻辑在这里
)

func RunApiServer(cfg *config.Config) {
	// 设置 Gin 模式
	if cfg.Server.DebugLevel == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	if cfg.Server.DebugLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// 使用内嵌的静态文件系统替代直接文件路径
	fs := assets.GetFileSystem()
	// 静态文件服务 - 使用内嵌资源
	r.StaticFS("/public", fs)

	// 提供根路径访问index.html
	r.GET("/", func(c *gin.Context) {
		// 尝试直接从内嵌文件系统读取index.html
		content, err := assets.GetFileContent("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	// 提供login.html访问
	r.GET("/login.html", func(c *gin.Context) {
		content, err := assets.GetFileContent("login.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load login.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	// 3. Session 设置 (对应 express-session + memorystore)
	// 使用 memstore (内存存储)，生产环境建议换成 redis
	secret := cfg.Auth.JavbusSessionSecret
	if secret == "" {
		secret = "_jav_bus_"
	}
	store := memstore.NewStore([]byte(secret))
	// 设置 Session 选项
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 1周 (秒)
		HttpOnly: true,
		Secure:   false, // 本地开发设为 false，HTTPS 环境设为 true
	})
	r.Use(sessions.Sessions("javbus.api.sid", store))

	// ==========================================
	// 公开 API (无需鉴权)
	// ==========================================

	// 判断是否启用了账号密码验证
	useCredentials := cfg.Admin.AdminUsername != "" && cfg.Admin.AdminPassword != ""

	// 获取用户信息
	r.GET("/api/user", func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		// 转换类型安全
		var userStr string
		if v, ok := username.(string); ok {
			userStr = v
		} else {
			// 如果 session 里是 nil，userStr 为 ""
		}

		// 注意：Go 中 nil 不会像 JS 那样自动变成 undefined，这里返回空字符串或不返回字段
		// 为了对齐前端 { username: string | undefined }，可以用指针
		var respUsername *string
		if userStr != "" {
			respUsername = &userStr
		}

		c.JSON(http.StatusOK, gin.H{
			"username":       respUsername,
			"useCredentials": useCredentials,
		})
	})

	// 登录接口
	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body"})
			return
		}

		// 验证用户名密码 (对应 loginValidators)
		// 注意：这里去除了 trim 等逻辑，Gin BindJSON 会原样解析，建议前端处理好或手动 Trim
		username := strings.TrimSpace(req.Username)
		password := strings.TrimSpace(req.Password)

		if !useCredentials || (username == cfg.Admin.AdminUsername && password == cfg.Admin.AdminPassword) {
			session := sessions.Default(c)
			session.Set("username", username)
			if err := session.Save(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Session save failed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Login success"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid username or password"})
		}
	})

	// 退出接口
	r.POST("/api/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()                               // 清空 session
		session.Options(sessions.Options{MaxAge: -1}) // 删除 cookie
		if err := session.Save(); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout success"})
	})

	// ==========================================
	// 受保护 API (需要鉴权)
	// ==========================================

	// 创建一个路由组，应用鉴权中间件
	// 对应 Node.js 中的 app.use((req, res, next) => { ... })
	api := r.Group("/api")
	api.Use(AuthMiddleware(cfg))
	{
		// 这里挂载原本 router.js 里的业务逻辑
		// 例如: handler.RegisterRoutes(api)
		// 假设你有这样的路由：
		// api.GET("/movies", handler.GetMovies)
		// api.GET("/movies/:id", handler.GetMovieDetail)
		RegisterRoutes(api, cfg)
	}

	// 404 处理 (NoRoute)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.ServerPort)
	log.Printf("Server starting on %s", addr)
	r.Run(addr)
}

// AuthMiddleware 核心鉴权逻辑
// 对应 Node.js 代码中间那大段 if (token) ... else if (useCredentials) ...
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.ReplaceAll(token, "Bearer ", "")
		session := sessions.Default(c)
		user := session.Get("username") // 之前存的是 "username"

		useCredentials := cfg.Admin.AdminUsername != "" && cfg.Admin.AdminPassword != ""
		originalURL := c.Request.URL.String()

		// 1. Token 验证优先
		if token != "" {
			if token == cfg.Auth.JavbusJwtToken {
				c.Next() // Token 正确，放行
				return
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
		}

		// 2. 账号密码 Session 验证
		if useCredentials {
			if user != nil {
				c.Next() // Session 存在，放行
				return
			} else {
				// 未登录，重定向
				// 注意：API 请求返回 redirect 可能会被前端 fetch 拦截或处理，视前端逻辑而定
				// 原代码逻辑是重定向到 /login.html
				target := "/login.html?redirect=" + url.QueryEscape(originalURL)
				c.Redirect(http.StatusFound, target)
				c.Abort()
				return
			}
		}

		// 3. 如果既没 Token 也没开启账号验证，直接放行
		c.Next()
	}
}
