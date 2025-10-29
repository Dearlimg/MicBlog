package repository

import (
	"blog/shared/models"
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// CommentRepository 评论仓库接口
type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id uint) (*models.Comment, error)
	GetCommentsByUserID(userID uint) ([]*models.Comment, error)
	GetAllComments() ([]*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uint) error
}

// commentRepository 评论仓库实现
type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建评论仓库
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.Comment{})
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

// CreateComment 创建评论
func (r *commentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// GetCommentByID 根据ID获取评论
func (r *commentRepository) GetCommentByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetCommentsByUserID 根据用户ID获取评论
func (r *commentRepository) GetCommentsByUserID(userID uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").Order("created_at DESC").Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// GetAllComments 获取所有评论
func (r *commentRepository) GetAllComments() ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.Where("status = ?", "active").Order("created_at DESC").Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// UpdateComment 更新评论
func (r *commentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

// DeleteComment 删除评论（软删除）
func (r *commentRepository) DeleteComment(id uint) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", id).Update("status", "deleted").Error
}
