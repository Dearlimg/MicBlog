package controller

import (
	"blog/shop-service/logic"
	"blog/shop-service/models"
	"strconv"

	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
)

// ProductController 商品控制器
type ProductController struct {
	productLogic *logic.ProductLogic
}

// OrderController 订单控制器
type OrderController struct {
	orderLogic *logic.OrderLogic
}

// CartController 购物车控制器
type CartController struct {
	cartLogic *logic.CartLogic
}

// NewProductController 创建商品控制器
func NewProductController(productLogic *logic.ProductLogic) *ProductController {
	return &ProductController{productLogic: productLogic}
}

// NewOrderController 创建订单控制器
func NewOrderController(orderLogic *logic.OrderLogic) *OrderController {
	return &OrderController{orderLogic: orderLogic}
}

// NewCartController 创建购物车控制器
func NewCartController(cartLogic *logic.CartLogic) *CartController {
	return &CartController{cartLogic: cartLogic}
}

// ========== Product Controller ==========

// CreateProduct 创建商品
func (pc *ProductController) CreateProduct(c *gin.Context) {
	rly := app.NewResponse(c)

	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	product, err := pc.productLogic.CreateProduct(&req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, product)
}

// GetProduct 获取商品详情
func (pc *ProductController) GetProduct(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid product ID"))
		return
	}

	product, err := pc.productLogic.GetProduct(uint(id))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, product)
}

// GetProducts 获取商品列表
func (pc *ProductController) GetProducts(c *gin.Context) {
	rly := app.NewResponse(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")

	products, total, err := pc.productLogic.GetProducts(page, pageSize, category)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, gin.H{
		"products":  products,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateProduct 更新商品
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid product ID"))
		return
	}

	var req models.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	product, err := pc.productLogic.UpdateProduct(uint(id), &req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, product)
}

// DeleteProduct 删除商品
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid product ID"))
		return
	}

	err = pc.productLogic.DeleteProduct(uint(id))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Product deleted successfully")
}

// ========== Order Controller ==========

// CreateOrder 创建订单
func (oc *OrderController) CreateOrder(c *gin.Context) {
	rly := app.NewResponse(c)

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	order, err := oc.orderLogic.CreateOrder(&req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, order)
}

// GetOrder 获取订单详情
func (oc *OrderController) GetOrder(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid order ID"))
		return
	}

	order, err := oc.orderLogic.GetOrder(uint(id))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, order)
}

// GetUserOrders 获取用户订单列表
func (oc *OrderController) GetUserOrders(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	orders, err := oc.orderLogic.GetUserOrders(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, orders)
}

// CancelOrder 取消订单
func (oc *OrderController) CancelOrder(c *gin.Context) {
	rly := app.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid order ID"))
		return
	}

	err = oc.orderLogic.CancelOrder(uint(id))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Order cancelled successfully")
}

// ========== Cart Controller ==========

// AddToCart 添加到购物车
func (cc *CartController) AddToCart(c *gin.Context) {
	rly := app.NewResponse(c)

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	err := cc.cartLogic.AddToCart(&req)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Item added to cart successfully")
}

// GetCart 获取购物车
func (cc *CartController) GetCart(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	cart, err := cc.cartLogic.GetCart(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, cart)
}

// UpdateCartItem 更新购物车商品数量
func (cc *CartController) UpdateCartItem(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	productIDStr := c.Param("product_id")

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid product ID"))
		return
	}

	var req struct {
		Quantity int `json:"quantity" validate:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	err = cc.cartLogic.UpdateCartItem(uint(userID), uint(productID), req.Quantity)
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Cart item updated successfully")
}

// RemoveFromCart 从购物车移除商品
func (cc *CartController) RemoveFromCart(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	productIDStr := c.Param("product_id")

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid product ID"))
		return
	}

	err = cc.cartLogic.RemoveFromCart(uint(userID), uint(productID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Item removed from cart successfully")
}

// ClearCart 清空购物车
func (cc *CartController) ClearCart(c *gin.Context) {
	rly := app.NewResponse(c)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid user ID"))
		return
	}

	err = cc.cartLogic.ClearCart(uint(userID))
	if err != nil {
		rly.Reply(errcode.ErrServer.WithDetails(err.Error()))
		return
	}

	rly.Reply(nil, "Cart cleared successfully")
}
