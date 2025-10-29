package controller

import (
	"blog/wallet-service/logic"
	"strconv"

	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
)

// WalletController 钱包控制器
type WalletController struct {
	walletLogic *logic.WalletLogic
}

// NewWalletController 创建钱包控制器
func NewWalletController(walletLogic *logic.WalletLogic) *WalletController {
	return &WalletController{walletLogic: walletLogic}
}

// CreateWallet 创建钱包
func (wc *WalletController) CreateWallet(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	wallet, err := wc.walletLogic.CreateWallet(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, wallet)
}

// GetWallet 获取钱包信息
func (wc *WalletController) GetWallet(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	wallet, err := wc.walletLogic.GetWallet(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, wallet)
}

// AddBalance 增加余额
func (wc *WalletController) AddBalance(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	var req struct {
		Amount      float64 `json:"amount" validate:"required,gt=0"`
		Description string  `json:"description" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	transaction, err := wc.walletLogic.AddBalance(uint(userID), req.Amount, req.Description)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, transaction)
}

// DeductBalance 扣除余额
func (wc *WalletController) DeductBalance(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	var req struct {
		Amount      float64 `json:"amount" validate:"required,gt=0"`
		Description string  `json:"description" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	transaction, err := wc.walletLogic.DeductBalance(uint(userID), req.Amount, req.Description)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, transaction)
}

// GetTransactions 获取交易记录
func (wc *WalletController) GetTransactions(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	transactions, err := wc.walletLogic.GetTransactions(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, transactions)
}

// Transfer 转账
func (wc *WalletController) Transfer(c *gin.Context) {
	rly := app.NewResponse(c)

	var req struct {
		FromUserID  uint    `json:"from_user_id" validate:"required"`
		ToUserID    uint    `json:"to_user_id" validate:"required"`
		Amount      float64 `json:"amount" validate:"required,gt=0"`
		Description string  `json:"description" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	err := wc.walletLogic.Transfer(req.FromUserID, req.ToUserID, req.Amount, req.Description)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Transfer completed successfully")
}

// Server HTTP服务器
type Server struct {
	router *gin.Engine
	port   string
}

// NewServer 创建HTTP服务器
func NewServer(port string, walletController *WalletController) *Server {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// 钱包相关路由
	api := router.Group("/api/v1")
	{
		wallets := api.Group("/wallets")
		{
			wallets.POST("/:user_id", walletController.CreateWallet)
			wallets.GET("/:user_id", walletController.GetWallet)
			wallets.POST("/:user_id/add", walletController.AddBalance)
			wallets.POST("/:user_id/deduct", walletController.DeductBalance)
			wallets.GET("/:user_id/transactions", walletController.GetTransactions)
			wallets.POST("/transfer", walletController.Transfer)
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
