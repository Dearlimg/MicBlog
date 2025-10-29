package repository

import (
	"blog/shared/models"
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// WalletRepository 钱包仓库接口
type WalletRepository interface {
	CreateWallet(wallet *models.Wallet) error
	GetWalletByUserID(userID uint) (*models.Wallet, error)
	UpdateWallet(wallet *models.Wallet) error
	SafeUpdateWalletBalance(userID uint, amount float64, operation string) (*models.Wallet, error)
	SafeTransfer(fromUserID, toUserID uint, amount float64, description string) error
}

// TransactionRepository 交易仓库接口
type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	GetTransactionByID(id uint) (*models.Transaction, error)
	GetTransactionsByUserID(userID uint) ([]*models.Transaction, error)
	GetTransactionsByProductID(productID uint) ([]*models.Transaction, error)
	GetTransactionsByOrderID(orderID uint) ([]*models.Transaction, error)
	UpdateTransaction(transaction *models.Transaction) error
}

// walletRepository 钱包仓库实现
type walletRepository struct {
	db *gorm.DB
}

// transactionRepository 交易仓库实现
type transactionRepository struct {
	db *gorm.DB
}

// NewWalletRepository 创建钱包仓库
func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

// NewTransactionRepository 创建交易仓库
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.Wallet{}, &models.Transaction{})
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

// CreateWallet 创建钱包
func (r *walletRepository) CreateWallet(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

// GetWalletByUserID 根据用户ID获取钱包
func (r *walletRepository) GetWalletByUserID(userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// UpdateWallet 更新钱包
func (r *walletRepository) UpdateWallet(wallet *models.Wallet) error {
	return r.db.Save(wallet).Error
}

// CreateTransaction 创建交易记录
func (r *transactionRepository) CreateTransaction(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

// GetTransactionByID 根据ID获取交易记录
func (r *transactionRepository) GetTransactionByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetTransactionsByUserID 根据用户ID获取交易记录
func (r *transactionRepository) GetTransactionsByUserID(userID uint) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionsByProductID 根据商品ID获取交易记录（为商城模块准备）
func (r *transactionRepository) GetTransactionsByProductID(productID uint) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Where("product_id = ?", productID).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionsByOrderID 根据订单ID获取交易记录（为商城模块准备）
func (r *transactionRepository) GetTransactionsByOrderID(orderID uint) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// UpdateTransaction 更新交易记录
func (r *transactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	return r.db.Save(transaction).Error
}

// SafeUpdateWalletBalance 安全的余额更新（使用数据库事务和锁）
func (r *walletRepository) SafeUpdateWalletBalance(userID uint, amount float64, operation string) (*models.Wallet, error) {
	var wallet models.Wallet

	// 使用事务和行锁
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 使用 SELECT FOR UPDATE 锁定行
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).
			First(&wallet).Error

		if err != nil {
			return err
		}

		// 检查扣除操作的余额
		if operation == "deduct" {
			if wallet.Balance < amount {
				return errors.New("insufficient balance")
			}
		}

		// 使用原子更新
		var updateAmount float64
		if operation == "add" {
			updateAmount = amount
		} else {
			updateAmount = -amount
		}

		// 原子更新余额（防止并发问题）
		result := tx.Model(&models.Wallet{}).
			Where("id = ? AND user_id = ?", wallet.ID, userID).
			Update("balance", gorm.Expr("balance + ?", updateAmount))

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("wallet not found or update failed")
		}

		// 重新读取更新后的余额
		err = tx.Where("id = ?", wallet.ID).First(&wallet).Error
		return err
	})

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// SafeTransfer 安全的转账（使用数据库事务）
func (r *walletRepository) SafeTransfer(fromUserID, toUserID uint, amount float64, description string, tx *gorm.DB) error {
	// 如果外部已经提供了事务，使用外部事务
	if tx != nil {
		return r.safeTransferInTx(tx, fromUserID, toUserID, amount, description)
	}

	// 否则创建新事务
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.safeTransferInTx(tx, fromUserID, toUserID, amount, description)
	})
}

// safeTransferInTx 在事务中执行转账
func (r *walletRepository) safeTransferInTx(tx *gorm.DB, fromUserID, toUserID uint, amount float64, description string) error {
	var fromWallet, toWallet models.Wallet

	// 锁定发送方钱包（按ID顺序锁定，防止死锁）
	if fromUserID < toUserID {
		// 先锁定fromWallet
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", fromUserID).
			First(&fromWallet).Error
		if err != nil {
			return err
		}

		// 检查余额
		if fromWallet.Balance < amount {
			return errors.New("insufficient balance")
		}

		// 再锁定toWallet
		err = tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", toUserID).
			First(&toWallet).Error
		if err != nil {
			// 如果接收方钱包不存在，创建它
			toWallet = models.Wallet{
				UserID:  toUserID,
				Balance: 0,
			}
			err = tx.Create(&toWallet).Error
			if err != nil {
				return err
			}
		}
	} else {
		// 反向锁定顺序
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", toUserID).
			First(&toWallet).Error
		if err != nil {
			toWallet = models.Wallet{
				UserID:  toUserID,
				Balance: 0,
			}
			err = tx.Create(&toWallet).Error
			if err != nil {
				return err
			}
		}

		err = tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", fromUserID).
			First(&fromWallet).Error
		if err != nil {
			return err
		}

		if fromWallet.Balance < amount {
			return errors.New("insufficient balance")
		}
	}

	// 原子扣除发送方余额（再次检查余额，乐观锁）
	result := tx.Model(&models.Wallet{}).
		Where("id = ? AND user_id = ? AND balance >= ?", fromWallet.ID, fromUserID, amount).
		Update("balance", gorm.Expr("balance - ?", amount))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("insufficient balance")
	}

	// 原子增加接收方余额
	err := tx.Model(&models.Wallet{}).
		Where("id = ?", toWallet.ID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
	if err != nil {
		return err
	}

	return nil
}
