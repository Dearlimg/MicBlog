package repository

import (
	"blog/shared/models"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
	CreateEmailVerification(verification *models.EmailVerification) error
	GetEmailVerification(email, code string) (*models.EmailVerification, error)
	DeleteEmailVerification(email string) error
}

// EmailRepository 邮箱仓库接口
type EmailRepository interface {
	SetVerificationCode(email, code string, expiration time.Duration) error
	GetVerificationCode(email string) (string, error)
	DeleteVerificationCode(email string) error
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// emailRepository 邮箱仓库实现
type emailRepository struct {
	redis *redis.Client
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// NewEmailRepository 创建邮箱仓库
func NewEmailRepository(redis *redis.Client) EmailRepository {
	return &emailRepository{redis: redis}
}

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.User{}, &models.EmailVerification{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// InitRedis 初始化Redis
func InitRedis(addr, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return rdb
}

// CreateUser 创建用户
func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail 根据邮箱获取用户
func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (r *userRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户
func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// CreateEmailVerification 创建邮箱验证记录
func (r *userRepository) CreateEmailVerification(verification *models.EmailVerification) error {
	return r.db.Create(verification).Error
}

// GetEmailVerification 获取邮箱验证记录
func (r *userRepository) GetEmailVerification(email, code string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.Where("email = ? AND code = ? AND expires_at > ?", email, code, time.Now()).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// DeleteEmailVerification 删除邮箱验证记录
func (r *userRepository) DeleteEmailVerification(email string) error {
	return r.db.Where("email = ?", email).Delete(&models.EmailVerification{}).Error
}

// SetVerificationCode 设置验证码
func (r *emailRepository) SetVerificationCode(email, code string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("verification:%s", email)
	return r.redis.Set(ctx, key, code, expiration).Err()
}

// GetVerificationCode 获取验证码
func (r *emailRepository) GetVerificationCode(email string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("verification:%s", email)
	return r.redis.Get(ctx, key).Result()
}

// DeleteVerificationCode 删除验证码
func (r *emailRepository) DeleteVerificationCode(email string) error {
	ctx := context.Background()
	key := fmt.Sprintf("verification:%s", email)
	return r.redis.Del(ctx, key).Err()
}
