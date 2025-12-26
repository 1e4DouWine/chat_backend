package router

import (
	"chat_backend/internal/middleware"
	v1 "chat_backend/internal/router/api/v1"

	"github.com/labstack/echo/v4"
)

// InitRouter 初始化所有路由
func InitRouter(e *echo.Echo) {
	// API v1 路由组
	apiV1 := e.Group("/api/v1")

	// 注册 hello 相关路由
	helloRoutes(apiV1)

	authRoutes(apiV1)
	userRoutes(apiV1)
}

// helloRoutes 注册hello相关路由
func helloRoutes(api *echo.Group) {
	api.GET("/hello", v1.Hello)
}

// authRoutes 注册认证相关路由
func authRoutes(api *echo.Group) {
	// 公共路由（不需要认证）
	auth := api.Group("/auth")
	auth.POST("/register", v1.Register)
	auth.POST("/login", v1.Login)
	auth.POST("/refresh", v1.RefreshToken)
}

// userRoutes 用户相关路由
func userRoutes(api *echo.Group) {
	user := api.Group("/user")
	user.Use(middleware.JWTMiddleware())
	user.GET("/me", v1.GetMe)
	user.GET("/friend", v1.GetFriendList)
	user.POST("/friend", v1.AddFriend)
	user.PUT("/friend/:id", v1.ProcessFriendRequest)
	user.DELETE("/friend/:id", v1.DeleteFriend)
}
