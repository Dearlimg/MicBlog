package logic

import (
	"blog/shared/email"
	"blog/shared/kafka"
	"blog/shared/models"
	"blog/user-service/repository"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/crypto/bcrypt"
)

// UserLogic 用户业务逻辑
type UserLogic struct {
	userRepo  repository.UserRepository
	emailRepo repository.EmailRepository
	producer  *kafka.Producer
	emailSvc  *email.EmailService
}

// NewUserLogic 创建用户业务逻辑
func NewUserLogic(userRepo repository.UserRepository, emailRepo repository.EmailRepository, producer *kafka.Producer, emailConfig email.EmailConfig) *UserLogic {
	emailSvc := email.NewEmailService(&emailConfig)
	return &UserLogic{
		userRepo:  userRepo,
		emailRepo: emailRepo,
		producer:  producer,
		emailSvc:  emailSvc,
	}
}

// Register 用户注册
func (ul *UserLogic) Register(req *models.UserRegisterRequest) (*models.UserResponse, error) {
	// 检查邮箱是否已存在
	existingUser, err := ul.userRepo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 创建用户
	user := &models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   string(hashedPassword),
		IsVerified: false,
	}

	err = ul.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// 发送验证邮件
	verificationCode := ul.generateVerificationCode()
	err = ul.emailSvc.SendVerificationEmail(user.Email, verificationCode)
	if err != nil {
		log.Printf("Failed to send verification email: %v", err)
		// 不返回错误，用户可以稍后重新发送
	}

	// 保存验证码到Redis
	err = ul.emailRepo.SetVerificationCode(user.Email, verificationCode, 10*time.Minute)
	if err != nil {
		log.Printf("Failed to save verification code: %v", err)
	}

	// 发送Kafka事件
	event := &models.UserRegisterEvent{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
	err = ul.producer.SendMessage(kafka.TopicUserRegister, fmt.Sprintf("%d", user.ID), event)
	if err != nil {
		log.Printf("Failed to send user register event: %v", err)
	}

	return &models.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt.Format(time.RFC3339),
	}, nil
}

// Login 用户登录
func (ul *UserLogic) Login(req *models.UserLoginRequest) (*models.UserResponse, error) {
	// 获取用户
	user, err := ul.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 发送Kafka事件
	event := &models.UserLoginEvent{
		UserID: user.ID,
		Email:  user.Email,
	}
	err = ul.producer.SendMessage(kafka.TopicUserLogin, fmt.Sprintf("%d", user.ID), event)
	if err != nil {
		log.Printf("Failed to send user login event: %v", err)
	}

	return &models.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt.Format(time.RFC3339),
	}, nil
}

// VerifyEmail 验证邮箱
func (ul *UserLogic) VerifyEmail(email, code string) error {
	// 从Redis获取验证码
	storedCode, err := ul.emailRepo.GetVerificationCode(email)
	if err != nil {
		return fmt.Errorf("verification code not found or expired")
	}

	if storedCode != code {
		return fmt.Errorf("invalid verification code")
	}

	// 获取用户
	user, err := ul.userRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// 更新用户验证状态
	user.IsVerified = true
	err = ul.userRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	// 删除验证码
	err = ul.emailRepo.DeleteVerificationCode(email)
	if err != nil {
		log.Printf("Failed to delete verification code: %v", err)
	}

	// 发送欢迎邮件
	err = ul.emailSvc.SendWelcomeEmail(user.Email, user.Username)
	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}

	return nil
}

// ResendVerificationCode 重新发送验证码
func (ul *UserLogic) ResendVerificationCode(email string) error {
	// 检查用户是否存在
	user, err := ul.userRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsVerified {
		return fmt.Errorf("email already verified")
	}

	// 生成新的验证码
	verificationCode := ul.generateVerificationCode()

	// 发送验证邮件
	err = ul.emailSvc.SendVerificationEmail(user.Email, verificationCode)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %v", err)
	}

	// 保存验证码到Redis
	err = ul.emailRepo.SetVerificationCode(user.Email, verificationCode, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to save verification code: %v", err)
	}

	return nil
}

// HandleEmailVerification 处理邮箱验证消息
func (ul *UserLogic) HandleEmailVerification(msg *sarama.ConsumerMessage) error {
	log.Printf("Received email verification message: %s", string(msg.Value))
	// 这里可以处理来自其他服务的邮箱验证请求
	return nil
}

// generateVerificationCode 生成验证码
func (ul *UserLogic) generateVerificationCode() string {
	// 生成6位数字验证码
	max := big.NewInt(999999)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

// GetUserProfile 获取用户信息
func (ul *UserLogic) GetUserProfile(userID uint) (*models.UserResponse, error) {
	user, err := ul.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &models.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt.Format(time.RFC3339),
	}, nil
}
