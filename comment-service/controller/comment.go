package controller

import (
	"blog/comment-service/logic"
	"blog/shared/models"
	"strconv"

	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
)

// CommentController 评论控制器
type CommentController struct {
	commentLogic *logic.CommentLogic
}

// NewCommentController 创建评论控制器
func NewCommentController(commentLogic *logic.CommentLogic) *CommentController {
	return &CommentController{commentLogic: commentLogic}
}

// CreateComment 创建评论
func (cc *CommentController) CreateComment(c *gin.Context) {
	rly := app.NewResponse(c)

	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	comment, err := cc.commentLogic.CreateComment(&req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, comment)
}

// GetComment 获取评论
func (cc *CommentController) GetComment(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid comment ID"))
		return
	}

	comment, err := cc.commentLogic.GetComment(uint(id))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, comment)
}

// GetComments 获取所有评论
func (cc *CommentController) GetComments(c *gin.Context) {
	rly := app.NewResponse(c)

	comments, err := cc.commentLogic.GetComments()
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.ReplyList(nil, int64(len(comments)), comments)
}

// GetCommentsByUser 获取用户评论
func (cc *CommentController) GetCommentsByUser(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	comments, err := cc.commentLogic.GetCommentsByUser(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.ReplyList(nil, int64(len(comments)), comments)
}

// UpdateComment 更新评论
func (cc *CommentController) UpdateComment(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid comment ID"))
		return
	}

	var req struct {
		Content string `json:"content" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	comment, err := cc.commentLogic.UpdateComment(uint(id), req.Content)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, comment)
}

// DeleteComment 删除评论
func (cc *CommentController) DeleteComment(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid comment ID"))
		return
	}

	err = cc.commentLogic.DeleteComment(uint(id))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Comment deleted successfully")
}

// HealthCheck 健康检查
func (cc *CommentController) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "comment-service",
	})
}

// Server HTTP服务器
type Server struct {
	router *gin.Engine
	port   string
}

// NewServer 创建HTTP服务器
func NewServer(port string, commentController *CommentController) *Server {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// 健康检查路由
	router.GET("/health", commentController.HealthCheck)

	// 评论相关路由
	api := router.Group("/api/v1")
	{
		comments := api.Group("/comments")
		{
			comments.POST("", commentController.CreateComment)
			comments.GET("/:id", commentController.GetComment)
			comments.GET("", commentController.GetComments)
			comments.GET("/user/:user_id", commentController.GetCommentsByUser)
			comments.PUT("/:id", commentController.UpdateComment)
			comments.DELETE("/:id", commentController.DeleteComment)
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
