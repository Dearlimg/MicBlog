package repository

import (
	"blog/shop-service/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProductRepository 商品仓库接口
type ProductRepository interface {
	CreateProduct(product *models.Product) error
	GetProductByID(id uint) (*models.Product, error)
	GetProducts(offset, limit int, category string) ([]*models.Product, int64, error)
	UpdateProduct(id uint, product *models.Product) error
	DeleteProduct(id uint) error
	ReduceStock(productID uint, quantity int) error
	IncreaseStock(productID uint, quantity int) error
}

// OrderRepository 订单仓库接口
type OrderRepository interface {
	CreateOrder(order *models.Order) error
	GetOrderByID(id uint) (*models.Order, error)
	GetOrdersByUserID(userID uint) ([]*models.Order, error)
	UpdateOrderStatus(id uint, status string) error
	CreateOrderItem(item *models.OrderItem) error
	GetOrderItemsByOrderID(orderID uint) ([]*models.OrderItem, error)
}

// CartRepository 购物车仓库接口
type CartRepository interface {
	AddToCart(cart *models.Cart) error
	GetCartByUserID(userID uint) ([]*models.Cart, error)
	UpdateCartItem(userID, productID uint, quantity int) error
	DeleteCartItem(userID, productID uint) error
	ClearCart(userID uint) error
	GetCartItem(userID, productID uint) (*models.Cart, error)
}

// productRepository 商品仓库实现
type productRepository struct {
	db *gorm.DB
}

// orderRepository 订单仓库实现
type orderRepository struct {
	db *gorm.DB
}

// cartRepository 购物车仓库实现
type cartRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓库
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// NewOrderRepository 创建订单仓库
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// NewCartRepository 创建购物车仓库
func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// ========== Product Repository Implementation ==========

func (r *productRepository) CreateProduct(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetProductByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetProducts(offset, limit int, category string) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Where("status = ?", "active")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&products).Error
	return products, total, err
}

func (r *productRepository) UpdateProduct(id uint, product *models.Product) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Updates(product).Error
}

func (r *productRepository) DeleteProduct(id uint) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("status", "offline").Error
}

func (r *productRepository) ReduceStock(productID uint, quantity int) error {
	return r.db.Model(&models.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity)).Error
}

func (r *productRepository) IncreaseStock(productID uint, quantity int) error {
	return r.db.Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

// ========== Order Repository Implementation ==========

func (r *orderRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetOrderByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("OrderItems").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetOrdersByUserID(userID uint) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.Where("user_id = ?", userID).Preload("OrderItems").Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateOrderStatus(id uint, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepository) CreateOrderItem(item *models.OrderItem) error {
	return r.db.Create(item).Error
}

func (r *orderRepository) GetOrderItemsByOrderID(orderID uint) ([]*models.OrderItem, error) {
	var items []*models.OrderItem
	err := r.db.Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// ========== Cart Repository Implementation ==========

func (r *cartRepository) AddToCart(cart *models.Cart) error {
	// 检查是否已存在
	var existingCart models.Cart
	err := r.db.Where("user_id = ? AND product_id = ?", cart.UserID, cart.ProductID).First(&existingCart).Error

	if err == nil {
		// 更新数量
		return r.db.Model(&existingCart).Update("quantity", gorm.Expr("quantity + ?", cart.Quantity)).Error
	}

	// 创建新记录
	return r.db.Create(cart).Error
}

func (r *cartRepository) GetCartByUserID(userID uint) ([]*models.Cart, error) {
	var carts []*models.Cart
	err := r.db.Where("user_id = ?", userID).Find(&carts).Error
	return carts, err
}

func (r *cartRepository) UpdateCartItem(userID, productID uint, quantity int) error {
	return r.db.Model(&models.Cart{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Update("quantity", quantity).Error
}

func (r *cartRepository) DeleteCartItem(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Cart{}).Error
}

func (r *cartRepository) ClearCart(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Cart{}).Error
}

func (r *cartRepository) GetCartItem(userID, productID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}
