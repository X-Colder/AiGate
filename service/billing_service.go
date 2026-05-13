package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/aigate/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{db: db}
}

func (s *BillingService) GetOrCreateBalance(userID, tenantID string) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := s.db.Where("user_id = ?", userID).First(&balance).Error
	if err == nil {
		return &balance, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	balance = model.UserBalance{
		ID:             uuid.New().String(),
		UserID:         userID,
		TenantID:       tenantID,
		Balance:        0,
		FreeBalance:    0,
		MonthlyResetAt: time.Now(),
	}
	if err := s.db.Create(&balance).Error; err != nil {
		return nil, err
	}
	return &balance, nil
}

func (s *BillingService) GetBalance(userID string) (*model.BalanceResponse, error) {
	var balance model.UserBalance
	if err := s.db.Where("user_id = ?", userID).First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.BalanceResponse{}, nil
		}
		return nil, err
	}
	return &model.BalanceResponse{
		Balance:        balance.Balance,
		FreeBalance:    balance.FreeBalance,
		TotalRecharged: balance.TotalRecharged,
		TotalConsumed:  balance.TotalConsumed,
		MonthlyUsed:    balance.MonthlyUsed,
	}, nil
}

func (s *BillingService) Recharge(req *model.RechargeRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var balance model.UserBalance
		if err := tx.Where("user_id = ?", req.UserID).First(&balance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				var user model.User
				if err := tx.Where("id = ?", req.UserID).First(&user).Error; err != nil {
					return fmt.Errorf("user not found")
				}
				balance = model.UserBalance{
					ID:             uuid.New().String(),
					UserID:         req.UserID,
					TenantID:       user.TenantID,
					MonthlyResetAt: time.Now(),
				}
				if err := tx.Create(&balance).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		balance.Balance += req.Amount
		balance.TotalRecharged += req.Amount
		if err := tx.Save(&balance).Error; err != nil {
			return err
		}
		desc := req.Description
		if desc == "" {
			desc = "管理员充值"
		}
		txn := model.BalanceTransaction{
			UserID:      req.UserID,
			TenantID:    balance.TenantID,
			Type:        "recharge",
			Amount:      req.Amount,
			Balance:     balance.Balance,
			Description: desc,
			CreatedAt:   time.Now(),
		}
		return tx.Create(&txn).Error
	})
}

func (s *BillingService) Deduct(userID, tenantID, relatedID string, amount float64, desc string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var balance model.UserBalance
		if err := tx.Where("user_id = ?", userID).First(&balance).Error; err != nil {
			return fmt.Errorf("balance not found")
		}
		if balance.FreeBalance >= amount {
			balance.FreeBalance -= amount
		} else if balance.FreeBalance > 0 {
			remaining := amount - balance.FreeBalance
			balance.FreeBalance = 0
			balance.Balance -= remaining
		} else {
			balance.Balance -= amount
		}
		balance.TotalConsumed += amount
		if err := tx.Save(&balance).Error; err != nil {
			return err
		}
		txn := model.BalanceTransaction{
			UserID:      userID,
			TenantID:    tenantID,
			Type:        "consume",
			Amount:      -amount,
			Balance:     balance.Balance + balance.FreeBalance,
			RelatedID:   relatedID,
			Description: desc,
			CreatedAt:   time.Now(),
		}
		return tx.Create(&txn).Error
	})
}

func (s *BillingService) GrantFreeQuota(userID, tenantID string, amount float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var balance model.UserBalance
		if err := tx.Where("user_id = ?", userID).First(&balance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				balance = model.UserBalance{
					ID:             uuid.New().String(),
					UserID:         userID,
					TenantID:       tenantID,
					MonthlyResetAt: time.Now(),
				}
				tx.Create(&balance)
			} else {
				return err
			}
		}
		balance.FreeBalance += amount
		if err := tx.Save(&balance).Error; err != nil {
			return err
		}
		txn := model.BalanceTransaction{
			UserID:      userID,
			TenantID:    tenantID,
			Type:        "free_grant",
			Amount:      amount,
			Balance:     balance.Balance + balance.FreeBalance,
			Description: "免费额度赠送",
			CreatedAt:   time.Now(),
		}
		return tx.Create(&txn).Error
	})
}

func (s *BillingService) GetTransactions(userID string, page, pageSize int) ([]model.BalanceTransaction, int64, error) {
	var total int64
	s.db.Model(&model.BalanceTransaction{}).Where("user_id = ?", userID).Count(&total)

	var txns []model.BalanceTransaction
	offset := (page - 1) * pageSize
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&txns).Error; err != nil {
		return nil, 0, err
	}
	return txns, total, nil
}

func (s *BillingService) CheckBalance(userID string, mc *model.ModelCatalog) error {
	var balance model.UserBalance
	if err := s.db.Where("user_id = ?", userID).First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("insufficient balance")
		}
		return err
	}
	switch mc.BillingMode {
	case "free_tier":
		if balance.FreeBalance <= 0 && balance.Balance <= 0 {
			return fmt.Errorf("free quota exhausted and no balance")
		}
	case "quota":
		if mc.MonthlyQuota > 0 && balance.MonthlyUsed >= mc.MonthlyQuota {
			return fmt.Errorf("monthly quota exceeded")
		}
	case "per_request":
		if balance.Balance+balance.FreeBalance < mc.RequestPrice {
			return fmt.Errorf("insufficient balance")
		}
	default:
		if balance.Balance+balance.FreeBalance <= 0 {
			return fmt.Errorf("insufficient balance")
		}
	}
	return nil
}

func (s *BillingService) CalculateCost(mc *model.ModelCatalog, inputTokens, outputTokens int64) float64 {
	switch mc.BillingMode {
	case "per_request":
		return mc.RequestPrice
	case "free_tier":
		return mc.InputPricePer1K*float64(inputTokens)/1000 + mc.OutputPricePer1K*float64(outputTokens)/1000
	case "quota":
		return 0
	default:
		return mc.InputPricePer1K*float64(inputTokens)/1000 + mc.OutputPricePer1K*float64(outputTokens)/1000
	}
}

func (s *BillingService) UpdateMonthlyUsage(userID string, tokens int64) error {
	return s.db.Model(&model.UserBalance{}).Where("user_id = ?", userID).
		Update("monthly_used", gorm.Expr("monthly_used + ?", tokens)).Error
}

func (s *BillingService) ListUserBalances() ([]model.UserBalanceDetail, error) {
	var results []model.UserBalanceDetail
	err := s.db.Table("user_balances").
		Select("user_balances.user_id, users.username, tenants.name as tenant_name, user_balances.balance, user_balances.free_balance, user_balances.total_recharged, user_balances.total_consumed").
		Joins("LEFT JOIN users ON users.id = user_balances.user_id").
		Joins("LEFT JOIN tenants ON tenants.id = user_balances.tenant_id").
		Scan(&results).Error
	return results, err
}
