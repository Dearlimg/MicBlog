package models

import (
	"time"
)

// Product 商品模型
type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:200;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Price       float64   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"default:0"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	Category    string    `json:"category" gorm:"size:50;index"`
	Status      string    `json:"status" gorm:"size:20;default:'active'"` // "active", "sold_out", "offline"
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Order 订单模型
type Order struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Status      string    `json:"status" gorm:"size:20;default:'pending'"` // "pending", "paid", "shipped", "completed", "cancelled", "refunded"
	TotalAmount float64   `json:"total_amount" gorm:"not null"`
	Address     string    `json:"address" gorm:"size:500"`
	Phone       string    `json:"phone" gorm:"size:20"`
	Remark      string    `json:"remark" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OrderItem 订单项模型
type OrderItem struct {
	ID          uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID     uint    `json:"order_id" gorm:"not null;index"`
	ProductID   uint    `json:"product_id" gorm:"not null;index"`
	Quantity    int     `json:"quantity" gorm:"not null"`
	UnitPrice   float64 `json:"unit_price" gorm:"not null"`
	TotalPrice  float64 `json:"total_price" gorm:"not null"`
	ProductName string  `json:"product_name" gorm:"size:200"` // 冗余字段，保存商品快照
}

// Cart 购物车模型
type Cart struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Quantity  int       `json:"quantity" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ProductCreateRequest 创建商品请求
type ProductCreateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

// ProductUpdateRequest 更新商品请求
type ProductUpdateRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock" validate:"omitempty,gte=0"`
	ImageURL    *string  `json:"image_url"`
	Category    *string  `json:"category"`
	Status      *string  `json:"status"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID  uint               `json:"user_id" validate:"required"`
	Items   []OrderItemRequest `json:"items" validate:"required,min=1"`
	Address string             `json:"address" validate:"required"`
	Phone   string             `json:"phone" validate:"required"`
	Remark  string             `json:"remark"`
	UseCart bool               `json:"use_cart"` // 是否使用购物车
}

// OrderItemRequest 订单项请求
type OrderItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	UserID    uint `json:"user_id" validate:"required"`
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

// UpdateCartRequest 更新购物车请求
type UpdateCartRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}
