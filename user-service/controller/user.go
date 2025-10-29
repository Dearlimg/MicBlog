package controller

import (
	"blog/shared/models"
	"blog/user-service/logic"
	"strconv"

	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	userLogic *logic.UserLogic
}

// NewUserController 创建用户控制器
func NewUserController(userLogic *logic.UserLogic) *UserController {
	return &UserController{userLogic: userLogic}
}

// Register 用户注册
func (uc *UserController) Register(c *gin.Context) {
	rly := app.NewResponse(c)

	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	user, err := uc.userLogic.Register(&req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, user)
}

// Login 用户登录
func (uc *UserController) Login(c *gin.Context) {
	rly := app.NewResponse(c)

	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	user, err := uc.userLogic.Login(&req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, user)
}

// VerifyEmail 验证邮箱
func (uc *UserController) VerifyEmail(c *gin.Context) {
	rly := app.NewResponse(c)

	var req struct {
		Email string `json:"email" validate:"required,email"`
		Code  string `json:"code" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	err := uc.userLogic.VerifyEmail(req.Email, req.Code)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Email verified successfully")
}

// ResendVerificationCode 重新发送验证码
func (uc *UserController) ResendVerificationCode(c *gin.Context) {
	rly := app.NewResponse(c)

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	err := uc.userLogic.ResendVerificationCode(req.Email)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Verification code sent successfully")
}

// GetUserProfile 获取用户信息
func (uc *UserController) GetUserProfile(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	user, err := uc.userLogic.GetUserProfile(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, user)
}

// Server HTTP服务器
type Server struct {
	router *gin.Engine
	port   string
}

// NewServer 创建HTTP服务器
func NewServer(port string, userController *UserController) *Server {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// 用户相关路由
	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("/register", userController.Register)
			users.POST("/login", userController.Login)
			users.POST("/verify-email", userController.VerifyEmail)
			users.POST("/resend-code", userController.ResendVerificationCode)
			users.GET("/:id", userController.GetUserProfile)
		}
	}

	return &Server{
		router: router,
		port:   port,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.router.Run(":" + s.port)
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	// 这里可以添加优雅关闭逻辑
}
