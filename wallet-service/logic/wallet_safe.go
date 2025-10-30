package logic

import (
	"blog/shared/kafka"
	"blog/shared/models"
	"blog/wallet-service/repository"
	"fmt"
	"log"
	"time"
)

// SafeWalletLogic 安全的钱包业务逻辑（支持高并发）
type SafeWalletLogic struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	producer        *kafka.Producer
}

// NewSafeWalletLogic 创建安全的钱包业务逻辑
func NewSafeWalletLogic(walletRepo repository.WalletRepository, transactionRepo repository.TransactionRepository, producer *kafka.Producer) *SafeWalletLogic {
	return &SafeWalletLogic{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		producer:        producer,
	}
}

// SafeAddBalance 安全的增加余额（原子操作）
func (swl *SafeWalletLogic) SafeAddBalance(userID uint, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// 使用安全的余额更新
	wallet, err := swl.walletRepo.SafeUpdateWalletBalance(userID, amount, "add")
	if err != nil {
		return nil, fmt.Errorf("failed to add balance: %v", err)
	}

	// 创建交易记录
	transaction := &models.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Type:        "income",
		Amount:      amount,
		Description: description,
		Status:      "completed",
		CreatedAt:   time.Now(),
	}

	err = swl.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		log.Printf("Failed to create transaction record: %v", err)
		// 交易已经完成，记录日志即可
	}

	// 异步发送Kafka事件（不影响主流程）
	if swl.producer != nil {
		go func() {
			event := &models.PaymentEvent{
				UserID:        userID,
				Amount:        amount,
				Description:   description,
				TransactionID: transaction.ID,
			}
			err := swl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("%d", transaction.ID), event)
			if err != nil {
				log.Printf("Failed to send payment event: %v", err)
			}
		}()
	}

	return transaction, nil
}

// SafeDeductBalance 安全的扣除余额（原子操作）
func (swl *SafeWalletLogic) SafeDeductBalance(userID uint, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// 使用安全的余额更新
	wallet, err := swl.walletRepo.SafeUpdateWalletBalance(userID, amount, "deduct")
	if err != nil {
		return nil, fmt.Errorf("failed to deduct balance: %v", err)
	}

	// 创建交易记录
	transaction := &models.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Type:        "expense",
		Amount:      amount,
		Description: description,
		Status:      "completed",
		CreatedAt:   time.Now(),
	}

	err = swl.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		log.Printf("Failed to create transaction record: %v", err)
	}

	// 异步发送Kafka事件
	if swl.producer != nil {
		go func() {
			event := &models.PaymentEvent{
				UserID:        userID,
				Amount:        -amount,
				Description:   description,
				TransactionID: transaction.ID,
			}
			err := swl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("%d", transaction.ID), event)
			if err != nil {
				log.Printf("Failed to send payment event: %v", err)
			}
		}()
	}

	return transaction, nil
}

// SafeTransfer 安全的转账（使用数据库事务保证原子性）
func (swl *SafeWalletLogic) SafeTransfer(fromUserID, toUserID uint, amount float64, description string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to yourself")
	}

	// 使用安全的转账方法（包含完整事务）
	err := swl.walletRepo.SafeTransfer(fromUserID, toUserID, amount, description)
	if err != nil {
		return fmt.Errorf("transfer failed: %v", err)
	}

	// 异步发送Kafka事件（转账已完成，记录日志即可）
	if swl.producer != nil {
		go func() {
			event := &models.PaymentEvent{
				UserID:        fromUserID,
				Amount:        -amount,
				Description:   fmt.Sprintf("Transfer to user %d: %s", toUserID, description),
				TransactionID: 0, // 事务中已创建
			}
			swl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("transfer-%d-%d", fromUserID, toUserID), event)
		}()
	}

	return nil
}
