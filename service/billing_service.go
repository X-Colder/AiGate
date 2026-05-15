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
	// If billing mode supports monthly subscription, check for an active one first
	if mc.BillingMode == "both" || mc.BillingMode == "monthly" {
		var sub model.ModelSubscription
		err := s.db.Where("user_id = ? AND model_catalog_id = ? AND status = 1 AND end_date > ?",
			userID, mc.ID, time.Now()).First(&sub).Error
		if err == nil {
			// Active subscription found — allow without token charge
			return nil
		}
	}

	// No active subscription or model is token-only — check token balance
	var balance model.UserBalance
	if err := s.db.Where("user_id = ?", userID).First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("insufficient balance")
		}
		return err
	}

	// If billing mode is monthly-only, subscription is required
	if mc.BillingMode == "monthly" {
		return fmt.Errorf("this model requires a monthly subscription")
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
		// "both", "prepaid", etc. — check token balance
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

// PurchaseSubscription creates a monthly subscription for a model
func (s *BillingService) PurchaseSubscription(userID, tenantID string, req *model.PurchaseSubscriptionRequest) (*model.SubscriptionResponse, error) {
	// 1. Look up model to get MonthlyPrice
	var mc model.ModelCatalog
	if err := s.db.Where("id = ?", req.ModelID).First(&mc).Error; err != nil {
		return nil, fmt.Errorf("model not found")
	}
	if mc.MonthlyPrice <= 0 {
		return nil, fmt.Errorf("this model does not support monthly subscription")
	}

	// 2. Determine paidBy
	paidBy := req.PaidBy
	if paidBy == "" {
		paidBy = "personal"
	}
	if paidBy == "team" && tenantID == "" {
		return nil, fmt.Errorf("team payment requires a tenant")
	}

	// 3. Check for existing active subscription
	var existing model.ModelSubscription
	err := s.db.Where("user_id = ? AND model_catalog_id = ? AND status = 1 AND end_date > ?",
		userID, mc.ID, time.Now()).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("you already have an active subscription for this model")
	}

	amount := mc.MonthlyPrice
	var sub model.ModelSubscription

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		deducted := false

		if paidBy == "team" {
			// Deduct from team admin's balance
			var tenant model.Tenant
			if err := tx.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
				return fmt.Errorf("tenant not found")
			}
			var teamBalance model.UserBalance
			if err := tx.Where("user_id = ?", tenant.OwnerID).First(&teamBalance).Error; err != nil {
				return fmt.Errorf("team admin balance not found")
			}
			if teamBalance.Balance+teamBalance.FreeBalance >= amount {
				if teamBalance.FreeBalance >= amount {
					teamBalance.FreeBalance -= amount
				} else if teamBalance.FreeBalance > 0 {
					remaining := amount - teamBalance.FreeBalance
					teamBalance.FreeBalance = 0
					teamBalance.Balance -= remaining
				} else {
					teamBalance.Balance -= amount
				}
				teamBalance.TotalConsumed += amount
				if err := tx.Save(&teamBalance).Error; err != nil {
					return err
				}
				txn := model.BalanceTransaction{
					UserID:      tenant.OwnerID,
					TenantID:    tenantID,
					Type:        "subscription",
					Amount:      -amount,
					Balance:     teamBalance.Balance + teamBalance.FreeBalance,
					Description: fmt.Sprintf("月租订阅 %s (团队付费, 用户 %s)", mc.Name, userID),
					CreatedAt:   time.Now(),
				}
				if err := tx.Create(&txn).Error; err != nil {
					return err
				}
				deducted = true
			} else {
				return fmt.Errorf("team balance insufficient")
			}
		} else {
			// Deduct from user's own balance
			var userBalance model.UserBalance
			if err := tx.Where("user_id = ?", userID).First(&userBalance).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Auto-fallback to team if possible
					if tenantID != "" {
						paidBy = "team"
					} else {
						return fmt.Errorf("insufficient balance")
					}
				} else {
					return err
				}
			} else if userBalance.Balance+userBalance.FreeBalance >= amount {
				if userBalance.FreeBalance >= amount {
					userBalance.FreeBalance -= amount
				} else if userBalance.FreeBalance > 0 {
					remaining := amount - userBalance.FreeBalance
					userBalance.FreeBalance = 0
					userBalance.Balance -= remaining
				} else {
					userBalance.Balance -= amount
				}
				userBalance.TotalConsumed += amount
				if err := tx.Save(&userBalance).Error; err != nil {
					return err
				}
				txn := model.BalanceTransaction{
					UserID:      userID,
					TenantID:    tenantID,
					Type:        "subscription",
					Amount:      -amount,
					Balance:     userBalance.Balance + userBalance.FreeBalance,
					Description: fmt.Sprintf("月租订阅 %s", mc.Name),
					CreatedAt:   time.Now(),
				}
				if err := tx.Create(&txn).Error; err != nil {
					return err
				}
				deducted = true
			} else if tenantID != "" {
				// Auto fallback to team balance
				paidBy = "team"
			} else {
				return fmt.Errorf("insufficient balance")
			}
		}

		// If personal was insufficient, fallback to team
		if !deducted && paidBy == "team" {
			var tenant model.Tenant
			if err := tx.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
				return fmt.Errorf("tenant not found")
			}
			var teamBalance model.UserBalance
			if err := tx.Where("user_id = ?", tenant.OwnerID).First(&teamBalance).Error; err != nil {
				return fmt.Errorf("team admin balance not found")
			}
			if teamBalance.Balance+teamBalance.FreeBalance < amount {
				return fmt.Errorf("insufficient balance (personal and team)")
			}
			if teamBalance.FreeBalance >= amount {
				teamBalance.FreeBalance -= amount
			} else if teamBalance.FreeBalance > 0 {
				remaining := amount - teamBalance.FreeBalance
				teamBalance.FreeBalance = 0
				teamBalance.Balance -= remaining
			} else {
				teamBalance.Balance -= amount
			}
			teamBalance.TotalConsumed += amount
			if err := tx.Save(&teamBalance).Error; err != nil {
				return err
			}
			txn := model.BalanceTransaction{
				UserID:      tenant.OwnerID,
				TenantID:    tenantID,
				Type:        "subscription",
				Amount:      -amount,
				Balance:     teamBalance.Balance + teamBalance.FreeBalance,
				Description: fmt.Sprintf("月租订阅 %s (团队付费, 用户 %s)", mc.Name, userID),
				CreatedAt:   time.Now(),
			}
			if err := tx.Create(&txn).Error; err != nil {
				return err
			}
		}

		// 4. Create ModelSubscription
		now := time.Now()
		sub = model.ModelSubscription{
			ID:             uuid.New().String(),
			UserID:         userID,
			TenantID:       tenantID,
			ModelCatalogID: mc.ID,
			PaidBy:         paidBy,
			StartDate:      now,
			EndDate:        now.AddDate(0, 0, 30),
			Amount:         amount,
			Status:         1,
			CreatedAt:      now,
		}
		return tx.Create(&sub).Error
	})
	if txErr != nil {
		return nil, txErr
	}

	return &model.SubscriptionResponse{
		ID:        sub.ID,
		ModelID:   mc.ID,
		ModelName: mc.Name,
		StartDate: sub.StartDate.Format("2006-01-02"),
		EndDate:   sub.EndDate.Format("2006-01-02"),
		PaidBy:    sub.PaidBy,
		Amount:    sub.Amount,
		Status:    sub.Status,
	}, nil
}

// GetActiveSubscription finds an active subscription for a user and model
func (s *BillingService) GetActiveSubscription(userID, modelID string) (*model.ModelSubscription, error) {
	var sub model.ModelSubscription
	err := s.db.Where("user_id = ? AND model_catalog_id = ? AND status = 1 AND end_date > ?",
		userID, modelID, time.Now()).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListSubscriptions returns all subscriptions for a user
func (s *BillingService) ListSubscriptions(userID string) ([]model.SubscriptionResponse, error) {
	var subs []model.ModelSubscription
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&subs).Error; err != nil {
		return nil, err
	}

	var results []model.SubscriptionResponse
	for _, sub := range subs {
		var mc model.ModelCatalog
		modelName := ""
		if err := s.db.Where("id = ?", sub.ModelCatalogID).First(&mc).Error; err == nil {
			modelName = mc.Name
		}
		results = append(results, model.SubscriptionResponse{
			ID:        sub.ID,
			ModelID:   sub.ModelCatalogID,
			ModelName: modelName,
			StartDate: sub.StartDate.Format("2006-01-02"),
			EndDate:   sub.EndDate.Format("2006-01-02"),
			PaidBy:    sub.PaidBy,
			Amount:    sub.Amount,
			Status:    sub.Status,
		})
	}
	return results, nil
}
