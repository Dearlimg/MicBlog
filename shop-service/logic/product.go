package logic

import (
	"blog/shop-service/models"
	"blog/shop-service/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ProductLogic 商品业务逻辑
type ProductLogic struct {
	productRepo repository.ProductRepository
}

// OrderLogic 订单业务逻辑
type OrderLogic struct {
	productRepo repository.ProductRepository
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	walletURL   string
}

// CartLogic 购物车业务逻辑
type CartLogic struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

// NewProductLogic 创建商品业务逻辑
func NewProductLogic(productRepo repository.ProductRepository) *ProductLogic {
	return &ProductLogic{productRepo: productRepo}
}

// NewOrderLogic 创建订单业务逻辑
func NewOrderLogic(productRepo repository.ProductRepository, orderRepo repository.OrderRepository, cartRepo repository.CartRepository, walletURL string) *OrderLogic {
	return &OrderLogic{
		productRepo: productRepo,
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		walletURL:   walletURL,
	}
}

// NewCartLogic 创建购物车业务逻辑
func NewCartLogic(cartRepo repository.CartRepository, productRepo repository.ProductRepository) *CartLogic {
	return &CartLogic{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// ========== Product Logic ==========

// CreateProduct 创建商品
func (pl *ProductLogic) CreateProduct(req *models.ProductCreateRequest) (*models.Product, error) {
	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
		Status:      "active",
	}

	err := pl.productRepo.CreateProduct(product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %v", err)
	}

	return product, nil
}

// GetProduct 获取商品详情
func (pl *ProductLogic) GetProduct(id uint) (*models.Product, error) {
	return pl.productRepo.GetProductByID(id)
}

// GetProducts 获取商品列表
func (pl *ProductLogic) GetProducts(page, pageSize int, category string) ([]*models.Product, int64, error) {
	offset := (page - 1) * pageSize
	return pl.productRepo.GetProducts(offset, pageSize, category)
}

// UpdateProduct 更新商品
func (pl *ProductLogic) UpdateProduct(id uint, req *models.ProductUpdateRequest) (*models.Product, error) {
	product, err := pl.productRepo.GetProductByID(id)
	if err != nil {
		return nil, fmt.Errorf("product not found: %v", err)
	}

	updateMap := make(map[string]interface{})
	if req.Name != nil {
		updateMap["name"] = *req.Name
	}
	if req.Description != nil {
		updateMap["description"] = *req.Description
	}
	if req.Price != nil {
		updateMap["price"] = *req.Price
	}
	if req.Stock != nil {
		updateMap["stock"] = *req.Stock
	}
	if req.ImageURL != nil {
		updateMap["image_url"] = *req.ImageURL
	}
	if req.Category != nil {
		updateMap["category"] = *req.Category
	}
	if req.Status != nil {
		updateMap["status"] = *req.Status
	}

	// 这里需要转换，先更新
	tempProduct := *product
	for k, v := range updateMap {
		switch k {
		case "name":
			tempProduct.Name = v.(string)
		case "description":
			tempProduct.Description = v.(string)
		case "price":
			tempProduct.Price = v.(float64)
		case "stock":
			tempProduct.Stock = v.(int)
		case "image_url":
			tempProduct.ImageURL = v.(string)
		case "category":
			tempProduct.Category = v.(string)
		case "status":
			tempProduct.Status = v.(string)
		}
	}

	err = pl.productRepo.UpdateProduct(id, &tempProduct)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %v", err)
	}

	return pl.productRepo.GetProductByID(id)
}

// DeleteProduct 删除商品（软删除）
func (pl *ProductLogic) DeleteProduct(id uint) error {
	return pl.productRepo.DeleteProduct(id)
}

// ========== Order Logic ==========

// CreateOrder 创建订单并支付
func (ol *OrderLogic) CreateOrder(req *models.CreateOrderRequest) (*models.Order, error) {
	// 计算订单总金额并验证商品
	var totalAmount float64
	var orderItems []*models.OrderItem

	// 如果使用购物车，从购物车获取商品
	if req.UseCart {
		cartItems, err := ol.cartRepo.GetCartByUserID(req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cart: %v", err)
		}

		if len(cartItems) == 0 {
			return nil, fmt.Errorf("cart is empty")
		}

		for _, cartItem := range cartItems {
			product, err := ol.productRepo.GetProductByID(cartItem.ProductID)
			if err != nil {
				return nil, fmt.Errorf("product %d not found: %v", cartItem.ProductID, err)
			}

			if product.Status != "active" {
				return nil, fmt.Errorf("product %d is not available", cartItem.ProductID)
			}

			if product.Stock < cartItem.Quantity {
				return nil, fmt.Errorf("product %d stock insufficient", cartItem.ProductID)
			}

			itemPrice := product.Price * float64(cartItem.Quantity)
			totalAmount += itemPrice

			orderItems = append(orderItems, &models.OrderItem{
				ProductID:   product.ID,
				Quantity:    cartItem.Quantity,
				UnitPrice:   product.Price,
				TotalPrice:  itemPrice,
				ProductName: product.Name,
			})
		}
	} else {
		// 从请求中获取商品
		for _, item := range req.Items {
			product, err := ol.productRepo.GetProductByID(item.ProductID)
			if err != nil {
				return nil, fmt.Errorf("product %d not found: %v", item.ProductID, err)
			}

			if product.Status != "active" {
				return nil, fmt.Errorf("product %d is not available", item.ProductID)
			}

			if product.Stock < item.Quantity {
				return nil, fmt.Errorf("product %d stock insufficient", item.ProductID)
			}

			itemPrice := product.Price * float64(item.Quantity)
			totalAmount += itemPrice

			orderItems = append(orderItems, &models.OrderItem{
				ProductID:   product.ID,
				Quantity:    item.Quantity,
				UnitPrice:   product.Price,
				TotalPrice:  itemPrice,
				ProductName: product.Name,
			})
		}
	}

	// 创建订单
	order := &models.Order{
		UserID:      req.UserID,
		Status:      "pending",
		TotalAmount: totalAmount,
		Address:     req.Address,
		Phone:       req.Phone,
		Remark:      req.Remark,
	}

	err := ol.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	// 创建订单项
	for _, item := range orderItems {
		item.OrderID = order.ID
		err = ol.orderRepo.CreateOrderItem(item)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %v", err)
		}

		// 扣减库存
		err = ol.productRepo.ReduceStock(item.ProductID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to reduce stock for product %d: %v", item.ProductID, err)
		}
	}

	// 调用钱包服务支付
	err = ol.payOrder(order.ID, req.UserID, totalAmount, orderItems)
	if err != nil {
		// 支付失败，回滚库存
		for _, item := range orderItems {
			ol.productRepo.IncreaseStock(item.ProductID, item.Quantity)
		}
		ol.orderRepo.UpdateOrderStatus(order.ID, "cancelled")
		return nil, fmt.Errorf("payment failed: %v", err)
	}

	// 支付成功，更新订单状态
	ol.orderRepo.UpdateOrderStatus(order.ID, "paid")

	// 如果使用购物车，清空购物车
	if req.UseCart {
		ol.cartRepo.ClearCart(req.UserID)
	}

	return order, nil
}

// payOrder 调用钱包服务支付订单
func (ol *OrderLogic) payOrder(orderID, userID uint, amount float64, items []*models.OrderItem) error {
	// 构建支付请求
	type PaymentReq struct {
		UserID    uint    `json:"user_id"`
		ProductID *uint   `json:"product_id"`
		Quantity  int     `json:"quantity"`
		UnitPrice float64 `json:"unit_price"`
		OrderID   *uint   `json:"order_id"`
	}

	// 如果有多个商品，需要分别支付（或者可以合并支付，这里简化处理）
	// 实际项目中，应该一次性支付总金额
	url := fmt.Sprintf("http://%s/api/v1/wallets/%d/deduct", ol.walletURL, userID)

	// 构建请求体（这里简化，实际应该发送所有商品信息）
	reqBody := map[string]interface{}{
		"amount":      amount,
		"description": fmt.Sprintf("Order #%d payment", orderID),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wallet service error: %s", string(body))
	}

	return nil
}

// GetOrder 获取订单详情
func (ol *OrderLogic) GetOrder(orderID uint) (*models.Order, error) {
	return ol.orderRepo.GetOrderByID(orderID)
}

// GetUserOrders 获取用户订单列表
func (ol *OrderLogic) GetUserOrders(userID uint) ([]*models.Order, error) {
	return ol.orderRepo.GetOrdersByUserID(userID)
}

// CancelOrder 取消订单
func (ol *OrderLogic) CancelOrder(orderID uint) error {
	order, err := ol.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	if order.Status != "pending" && order.Status != "paid" {
		return fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

	// 恢复库存
	items, err := ol.orderRepo.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %v", err)
	}

	for _, item := range items {
		ol.productRepo.IncreaseStock(item.ProductID, item.Quantity)
	}

	// 更新订单状态
	return ol.orderRepo.UpdateOrderStatus(orderID, "cancelled")
}

// ========== Cart Logic ==========

// AddToCart 添加到购物车
func (cl *CartLogic) AddToCart(req *models.AddToCartRequest) error {
	// 检查商品是否存在
	product, err := cl.productRepo.GetProductByID(req.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %v", err)
	}

	if product.Status != "active" {
		return fmt.Errorf("product is not available")
	}

	cart := &models.Cart{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	return cl.cartRepo.AddToCart(cart)
}

// GetCart 获取购物车
func (cl *CartLogic) GetCart(userID uint) ([]*models.Cart, error) {
	return cl.cartRepo.GetCartByUserID(userID)
}

// UpdateCartItem 更新购物车商品数量
func (cl *CartLogic) UpdateCartItem(userID, productID uint, quantity int) error {
	return cl.cartRepo.UpdateCartItem(userID, productID, quantity)
}

// RemoveFromCart 从购物车移除商品
func (cl *CartLogic) RemoveFromCart(userID, productID uint) error {
	return cl.cartRepo.DeleteCartItem(userID, productID)
}

// ClearCart 清空购物车
func (cl *CartLogic) ClearCart(userID uint) error {
	return cl.cartRepo.ClearCart(userID)
}
