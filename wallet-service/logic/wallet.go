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

// PurchaseProduct 购买商品（为商城模块准备）
func (wl *WalletLogic) PurchaseProduct(userID uint, productID uint, quantity int, unitPrice float64, orderID *uint) (*models.Transaction, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}
	if unitPrice <= 0 {
		return nil, fmt.Errorf("unit price must be greater than 0")
	}

	totalAmount := float64(quantity) * unitPrice

	// 获取钱包
	wallet, err := wl.GetWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %v", err)
	}

	// 检查余额是否足够
	if wallet.Balance < totalAmount {
		return nil, fmt.Errorf("insufficient balance: need %.2f, have %.2f", totalAmount, wallet.Balance)
	}

	// 创建交易记录
	transaction := &models.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Type:        "purchase",
		Amount:      totalAmount,
		Description: fmt.Sprintf("Purchase product %d, quantity: %d", productID, quantity),
		ProductID:   &productID,
		OrderID:     orderID,
		Quantity:    quantity,
		Status:      "pending",
	}

	err = wl.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	// 扣除余额
	wallet.Balance -= totalAmount
	err = wl.walletRepo.UpdateWallet(wallet)
	if err != nil {
		// 回滚交易状态
		transaction.Status = "failed"
		wl.transactionRepo.UpdateTransaction(transaction)
		return nil, fmt.Errorf("failed to deduct balance: %v", err)
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
		Amount:        -totalAmount, // 负数表示支出
		Description:   transaction.Description,
		TransactionID: transaction.ID,
		ProductID:     &productID,
		OrderID:       orderID,
		Quantity:      quantity,
	}
	err = wl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("%d", transaction.ID), event)
	if err != nil {
		log.Printf("Failed to send payment event: %v", err)
	}

	return transaction, nil
}

// Refund 退款（为商城模块准备）
func (wl *WalletLogic) Refund(transactionID uint, reason string) (*models.Transaction, error) {
	// 获取原始交易记录
	originalTransaction, err := wl.transactionRepo.GetTransactionByID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %v", err)
	}

	// 检查是否已经退款过
	if originalTransaction.Status == "refunded" {
		return nil, fmt.Errorf("transaction already refunded")
	}

	// 只允许退款购买类型的交易
	if originalTransaction.Type != "purchase" {
		return nil, fmt.Errorf("only purchase transactions can be refunded")
	}

	// 获取钱包
	wallet, err := wl.GetWallet(originalTransaction.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %v", err)
	}

	// 创建退款交易记录
	refundTransaction := &models.Transaction{
		UserID:      originalTransaction.UserID,
		WalletID:    originalTransaction.WalletID,
		Type:        "refund",
		Amount:      originalTransaction.Amount, // 退款金额等于原交易金额
		Description: fmt.Sprintf("Refund for transaction %d: %s", transactionID, reason),
		ProductID:   originalTransaction.ProductID, // 关联原商品
		OrderID:     originalTransaction.OrderID,   // 关联原订单
		Quantity:    originalTransaction.Quantity,
		Status:      "pending",
	}

	err = wl.transactionRepo.CreateTransaction(refundTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %v", err)
	}

	// 退回余额
	wallet.Balance += originalTransaction.Amount
	err = wl.walletRepo.UpdateWallet(wallet)
	if err != nil {
		// 回滚退款交易状态
		refundTransaction.Status = "failed"
		wl.transactionRepo.UpdateTransaction(refundTransaction)
		return nil, fmt.Errorf("failed to refund balance: %v", err)
	}

	// 更新退款交易状态
	refundTransaction.Status = "completed"
	err = wl.transactionRepo.UpdateTransaction(refundTransaction)
	if err != nil {
		log.Printf("Failed to update refund transaction status: %v", err)
	}

	// 更新原交易状态为已退款
	originalTransaction.Status = "refunded"
	err = wl.transactionRepo.UpdateTransaction(originalTransaction)
	if err != nil {
		log.Printf("Failed to update original transaction status: %v", err)
	}

	// 发送Kafka退款事件
	event := &models.PaymentEvent{
		UserID:        originalTransaction.UserID,
		Amount:        originalTransaction.Amount, // 正数表示退款收入
		Description:   refundTransaction.Description,
		TransactionID: refundTransaction.ID,
		ProductID:     originalTransaction.ProductID,
		OrderID:       originalTransaction.OrderID,
		Quantity:      originalTransaction.Quantity,
	}
	err = wl.producer.SendMessage(kafka.TopicWalletPayment, fmt.Sprintf("refund-%d", refundTransaction.ID), event)
	if err != nil {
		log.Printf("Failed to send refund event: %v", err)
	}

	return refundTransaction, nil
}

// GetTransactionsByProductID 根据商品ID获取交易记录（为商城模块准备）
func (wl *WalletLogic) GetTransactionsByProductID(productID uint) ([]*models.Transaction, error) {
	// 这里需要在repository中添加查询方法
	// 暂时返回所有交易，由调用方过滤
	allTransactions, err := wl.transactionRepo.GetTransactionsByUserID(0) // 0表示获取所有
	if err != nil {
		return nil, err
	}

	var result []*models.Transaction
	for _, t := range allTransactions {
		if t.ProductID != nil && *t.ProductID == productID {
			result = append(result, t)
		}
	}
	return result, nil
}
