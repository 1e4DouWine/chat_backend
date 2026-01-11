package router

import (
	"chat_backend/internal/middleware"
	v1 "chat_backend/internal/router/api/v1"
	"chat_backend/internal/websocket"

	"github.com/labstack/echo/v4"
)

// InitRouter 初始化所有路由
func InitRouter(e *echo.Echo, rateLimiter *middleware.RateLimiter) {
	// API v1 路由组
	apiV1 := e.Group("/api/v1")

	// 注册 hello 相关路由
	helloRoutes(apiV1)

	authRoutes(apiV1, rateLimiter)
	userRoutes(apiV1, rateLimiter)
	groupRoutes(apiV1, rateLimiter)
	messageRoutes(apiV1, rateLimiter)
	wsRoutes(e)
}

// helloRoutes 注册hello相关路由
func helloRoutes(api *echo.Group) {
	api.GET("/hello", v1.Hello)
}

// authRoutes 注册认证相关路由
func authRoutes(api *echo.Group, rateLimiter *middleware.RateLimiter) {
	// 公共路由（不需要认证）
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimitMiddleware(rateLimiter, middleware.AuthRateLimit))
	auth.POST("/register", v1.Register)
	auth.POST("/login", v1.Login)
	auth.POST("/refresh", v1.RefreshToken)
}

// userRoutes 用户相关路由
func userRoutes(api *echo.Group, rateLimiter *middleware.RateLimiter) {
	user := api.Group("/user")
	user.Use(middleware.JWTMiddleware())
	user.Use(middleware.RateLimitMiddleware(rateLimiter, middleware.GeneralRateLimit))
	user.GET("/me", v1.GetMe)

	user.GET("/search", v1.SearchUser)

	user.GET("/friend", v1.GetFriendList)
	user.POST("/friend", v1.AddFriend)
	user.PUT("/friend/:id", v1.ProcessFriendRequest)
	user.DELETE("/friend/:id", v1.DeleteFriend)
}

// groupRoutes 群组相关路由
func groupRoutes(api *echo.Group, rateLimiter *middleware.RateLimiter) {
	group := api.Group("/group")
	group.Use(middleware.JWTMiddleware())
	group.Use(middleware.RateLimitMiddleware(rateLimiter, middleware.GeneralRateLimit))

	// 创建群组
	group.POST("", v1.CreateGroup)

	// 获取群组列表
	group.GET("", v1.GetGroupList)

	// 获取群组详情
	group.GET("/:id", v1.GetGroupDetail)

	// 搜索群组
	group.GET("/search", v1.SearchGroup)

	// 申请加入群组
	group.POST("/:id/request-join", v1.RequestJoinGroup)

	// 获取待审核的入群请求
	group.GET("/join-requests", v1.GetPendingJoinRequests)

	// 审批入群请求
	group.POST("/:id/join-requests/:user_id/approve", v1.ApproveJoinRequest)

	// 加入群组
	//group.POST("/:id/join", v1.JoinGroup)

	// 通过邀请码加入群组
	//group.POST("/join-by-code", v1.JoinGroupByCode)

	// 退出群组
	group.POST("/:id/leave", v1.LeaveGroup)

	// 解散群组
	group.DELETE("/:id", v1.DisbandGroup)

	// 转让群组
	group.PUT("/:id/transfer", v1.TransferGroup)

	// 移除群组成员
	group.DELETE("/:group_id/member/:user_id", v1.RemoveMember)
}

// wsRoutes WebSocket相关路由
func wsRoutes(e *echo.Echo) {
	ws := e.Group("/ws")
	ws.Use(middleware.JWTMiddleware())
	ws.GET("", websocket.HandleWebSocket)

	// WebSocket 状态查询API
	wsAPI := e.Group("/api/v1/ws")
	wsAPI.Use(middleware.JWTMiddleware())
	wsAPI.GET("/online", websocket.GetOnlineUsers)
	wsAPI.GET("/online/:id", websocket.IsUserOnline)
}

// messageRoutes 消息相关路由
func messageRoutes(api *echo.Group, rateLimiter *middleware.RateLimiter) {
	message := api.Group("/message")
	message.Use(middleware.JWTMiddleware())
	message.Use(middleware.RateLimitMiddleware(rateLimiter, middleware.MessageRateLimit))

	// 获取会话列表
	message.GET("/conversations", v1.GetConversationList)

	// 获取私聊消息记录
	message.GET("/private", v1.GetPrivateMessages)

	// 获取群聊消息记录
	message.GET("/group/:id", v1.GetGroupMessages)
}
