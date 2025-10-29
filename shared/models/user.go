package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username   string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Email      string    `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Password   string    `json:"-" gorm:"size:255;not null"`
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
	CreatedAt  string `json:"created_at"`
}

// EmailVerification 邮箱验证
type EmailVerification struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string    `json:"email" gorm:"size:100;not null"`
	Code      string    `json:"code" gorm:"size:10;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Wallet 钱包模型
type Wallet struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Balance   float64   `json:"balance" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Transaction 交易记录
type Transaction struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	WalletID    uint      `json:"wallet_id" gorm:"not null;index"`
	Type        string    `json:"type" gorm:"size:20;not null"` // "income", "expense"
	Amount      float64   `json:"amount" gorm:"not null"`
	Description string    `json:"description" gorm:"size:255"`
	Status      string    `json:"status" gorm:"size:20;default:'pending'"` // "pending", "completed", "failed"
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	UserID      uint    `json:"user_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description" validate:"required"`
}

// Comment 评论模型
type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	ParentID  *uint     `json:"parent_id"`                              // 父评论ID，用于回复
	Status    string    `json:"status" gorm:"size:20;default:'active'"` // "active", "deleted"
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CommentRequest 评论请求
type CommentRequest struct {
	UserID   uint   `json:"user_id" validate:"required"`
	Content  string `json:"content" validate:"required"`
	ParentID *uint  `json:"parent_id"`
}

// CommentResponse 评论响应
type CommentResponse struct {
	ID        uint              `json:"id"`
	UserID    uint              `json:"user_id"`
	Username  string            `json:"username"`
	Content   string            `json:"content"`
	ParentID  *uint             `json:"parent_id"`
	Status    string            `json:"status"`
	CreatedAt string            `json:"created_at"`
	Replies   []CommentResponse `json:"replies,omitempty"`
}

// KafkaMessage Kafka消息结构
type KafkaMessage struct {
	Type      string      `json:"type"`
	Service   string      `json:"service"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// UserRegisterEvent 用户注册事件
type UserRegisterEvent struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserLoginEvent 用户登录事件
type UserLoginEvent struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}

// PaymentEvent 支付事件
type PaymentEvent struct {
	UserID        uint    `json:"user_id"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	TransactionID uint    `json:"transaction_id"`
}

// CommentEvent 评论事件
type CommentEvent struct {
	CommentID uint   `json:"comment_id"`
	UserID    uint   `json:"user_id"`
	Content   string `json:"content"`
	Action    string `json:"action"` // "create", "update", "delete"
}
