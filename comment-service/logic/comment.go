package logic

import (
	"blog/comment-service/repository"
	"blog/shared/kafka"
	"blog/shared/models"
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
)

// CommentLogic 评论业务逻辑
type CommentLogic struct {
	commentRepo repository.CommentRepository
	producer    *kafka.Producer
}

// NewCommentLogic 创建评论业务逻辑
func NewCommentLogic(commentRepo repository.CommentRepository, producer *kafka.Producer) *CommentLogic {
	return &CommentLogic{
		commentRepo: commentRepo,
		producer:    producer,
	}
}

// CreateComment 创建评论
func (cl *CommentLogic) CreateComment(req *models.CommentRequest) (*models.CommentResponse, error) {
	// 创建评论
	comment := &models.Comment{
		UserID:   req.UserID,
		Content:  req.Content,
		ParentID: req.ParentID,
		Status:   "active",
	}

	err := cl.commentRepo.CreateComment(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %v", err)
	}

	// 发送Kafka事件
	event := &models.CommentEvent{
		CommentID: comment.ID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		Action:    "create",
	}
	err = cl.producer.SendMessage(kafka.TopicCommentCreate, fmt.Sprintf("%d", comment.ID), event)
	if err != nil {
		log.Printf("Failed to send comment event: %v", err)
	}

	// 构建响应
	response := &models.CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		ParentID:  comment.ParentID,
		Status:    comment.Status,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetComment 获取评论
func (cl *CommentLogic) GetComment(id uint) (*models.CommentResponse, error) {
	comment, err := cl.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, fmt.Errorf("comment not found")
	}

	response := &models.CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		ParentID:  comment.ParentID,
		Status:    comment.Status,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetComments 获取所有评论
func (cl *CommentLogic) GetComments() ([]*models.CommentResponse, error) {
	comments, err := cl.commentRepo.GetAllComments()
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}

	responses := make([]*models.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		response := &models.CommentResponse{
			ID:        comment.ID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			ParentID:  comment.ParentID,
			Status:    comment.Status,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetCommentsByUser 获取用户评论
func (cl *CommentLogic) GetCommentsByUser(userID uint) ([]*models.CommentResponse, error) {
	comments, err := cl.commentRepo.GetCommentsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user comments: %v", err)
	}

	responses := make([]*models.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		response := &models.CommentResponse{
			ID:        comment.ID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			ParentID:  comment.ParentID,
			Status:    comment.Status,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// UpdateComment 更新评论
func (cl *CommentLogic) UpdateComment(id uint, content string) (*models.CommentResponse, error) {
	comment, err := cl.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, fmt.Errorf("comment not found")
	}

	comment.Content = content
	err = cl.commentRepo.UpdateComment(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %v", err)
	}

	// 发送Kafka事件
	event := &models.CommentEvent{
		CommentID: comment.ID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		Action:    "update",
	}
	err = cl.producer.SendMessage(kafka.TopicCommentUpdate, fmt.Sprintf("%d", comment.ID), event)
	if err != nil {
		log.Printf("Failed to send comment event: %v", err)
	}

	response := &models.CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		ParentID:  comment.ParentID,
		Status:    comment.Status,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// DeleteComment 删除评论
func (cl *CommentLogic) DeleteComment(id uint) error {
	comment, err := cl.commentRepo.GetCommentByID(id)
	if err != nil {
		return fmt.Errorf("comment not found")
	}

	err = cl.commentRepo.DeleteComment(id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %v", err)
	}

	// 发送Kafka事件
	event := &models.CommentEvent{
		CommentID: comment.ID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		Action:    "delete",
	}
	err = cl.producer.SendMessage(kafka.TopicCommentDelete, fmt.Sprintf("%d", comment.ID), event)
	if err != nil {
		log.Printf("Failed to send comment event: %v", err)
	}

	return nil
}

// HandleCommentEvent 处理评论事件
func (cl *CommentLogic) HandleCommentEvent(msg *sarama.ConsumerMessage) error {
	log.Printf("Received comment event: %s", string(msg.Value))
	// 这里可以处理来自其他服务的评论事件
	return nil
}
