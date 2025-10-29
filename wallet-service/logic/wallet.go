package logic

import (
	"blog/shared/kafka"
	"blog/shared/models"
	"blog/wallet-service/repository"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

// WalletLogic 钱包业务逻辑
type WalletLogic struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	producer        *kafka.Producer
}

// NewWalletLogic 创建钱包业务逻辑
func NewWalletLogic(walletRepo repository.WalletRepository, transactionRepo repository.TransactionRepository, producer *kafka.Producer) *WalletLogic {
	return &WalletLogic{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		producer:        producer,
	}
}

// CreateWallet 创建钱包
func (wl *WalletLogic) CreateWallet(userID uint) (*models.Wallet, error) {
	// 检查用户是否已有钱包
	existingWallet, err := wl.walletRepo.GetWalletByUserID(userID)
	if err == nil && existingWallet != nil {
		return existingWallet, nil
	}

	// 创建新钱包
	wallet := &models.Wallet{
		UserID:  userID,
		Balance: 0.0,
	}

	err = wl.walletRepo.CreateWallet(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	return wallet, nil
}

// GetWallet 获取钱包信息
func (wl *WalletLogic) GetWallet(userID uint) (*models.Wallet, error) {
	wallet, err := wl.walletRepo.GetWalletByUserID(userID)
	if err != nil {
		// 如果钱包不存在，创建一个
		return wl.CreateWallet(userID)
	}
	return wallet, nil
}

// AddBalance 增加余额
func (wl *WalletLogic) AddBalance(userID uint, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// 获取钱包
	wallet, err := wl.GetWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %v", err)
	}

	// 创建交易记录
	transaction := &models.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Type:        "income",
		Amount:      amount,
		Description: description,
		Status:      "pending",
	}

	err = wl.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	// 更新钱包余额
	wallet.Balance += amount
	err = wl.walletRepo.UpdateWallet(wallet)
	if err != nil {
		// 回滚交易状态
		transaction.Status = "failed"
		wl.transactionRepo.UpdateTransaction(transaction)
		return nil, fmt.Errorf("failed to update wallet: %v", err)
	}

	// 更新交易状态
	transaction.Status = "completed"
	err = wl.transactionRepo.UpdateTransaction(transaction)
	if err != nil {
		log.Printf("Failed to update transaction status: %v", err)
	}

	// 发送Kafka事件
	event := &models.PaymentEvent{
		UserID:        userID,
		Amount:        amount,
		Description:   description,
		TransactionID: transaction.ID,
	}
	err = wl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("%d", transaction.ID), event)
	if err != nil {
		log.Printf("Failed to send payment event: %v", err)
	}

	return transaction, nil
}

// DeductBalance 扣除余额
func (wl *WalletLogic) DeductBalance(userID uint, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// 获取钱包
	wallet, err := wl.GetWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %v", err)
	}

	// 检查余额是否足够
	if wallet.Balance < amount {
		return nil, fmt.Errorf("insufficient balance")
	}

	// 创建交易记录
	transaction := &models.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Type:        "expense",
		Amount:      amount,
		Description: description,
		Status:      "pending",
	}

	err = wl.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	// 更新钱包余额
	wallet.Balance -= amount
	err = wl.walletRepo.UpdateWallet(wallet)
	if err != nil {
		// 回滚交易状态
		transaction.Status = "failed"
		wl.transactionRepo.UpdateTransaction(transaction)
		return nil, fmt.Errorf("failed to update wallet: %v", err)
	}

	// 更新交易状态
	transaction.Status = "completed"
	err = wl.transactionRepo.UpdateTransaction(transaction)
	if err != nil {
		log.Printf("Failed to update transaction status: %v", err)
	}

	// 发送Kafka事件
	event := &models.PaymentEvent{
		UserID:        userID,
		Amount:        -amount, // 负数表示支出
		Description:   description,
		TransactionID: transaction.ID,
	}
	err = wl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("%d", transaction.ID), event)
	if err != nil {
		log.Printf("Failed to send payment event: %v", err)
	}

	return transaction, nil
}

// GetTransactions 获取交易记录
func (wl *WalletLogic) GetTransactions(userID uint) ([]*models.Transaction, error) {
	return wl.transactionRepo.GetTransactionsByUserID(userID)
}

// HandlePaymentEvent 处理支付事件
func (wl *WalletLogic) HandlePaymentEvent(msg *sarama.ConsumerMessage) error {
	log.Printf("Received payment event: %s", string(msg.Value))
	// 这里可以处理来自其他服务的支付事件
	return nil
}

// Transfer 转账
func (wl *WalletLogic) Transfer(fromUserID, toUserID uint, amount float64, description string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to yourself")
	}

	// 扣除发送方余额
	_, err := wl.DeductBalance(fromUserID, amount, fmt.Sprintf("Transfer to user %d: %s", toUserID, description))
	if err != nil {
		return fmt.Errorf("failed to deduct balance: %v", err)
	}

	// 增加接收方余额
	_, err = wl.AddBalance(toUserID, amount, fmt.Sprintf("Transfer from user %d: %s", fromUserID, description))
	if err != nil {
		// 如果增加失败，需要回滚扣除操作
		log.Printf("Failed to add balance, transaction may be inconsistent: %v", err)
		return fmt.Errorf("failed to add balance: %v", err)
	}

	return nil
}
