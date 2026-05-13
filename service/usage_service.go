package service

import (
	"time"

	"github.com/aigate/model"
	"gorm.io/gorm"
)

type UsageService struct {
	db *gorm.DB
}

func NewUsageService(db *gorm.DB) *UsageService {
	return &UsageService{db: db}
}

func (s *UsageService) Record(record *model.UsageRecord) error {
	return s.db.Create(record).Error
}

func (s *UsageService) GetSummary(userID string, query *model.UsageQuery) (*model.UsageSummary, error) {
	startTime, _ := time.Parse("2006-01-02", query.StartDate)
	endTime, _ := time.Parse("2006-01-02", query.EndDate)
	endTime = endTime.Add(24 * time.Hour)

	tx := s.db.Model(&model.UsageRecord{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startTime, endTime)
	if query.ModelID != "" {
		tx = tx.Where("model_catalog_id = ?", query.ModelID)
	}
	if query.APIKeyID != "" {
		tx = tx.Where("api_key_id = ?", query.APIKeyID)
	}

	var summary model.UsageSummary
	tx.Select("COUNT(*) as total_requests, COALESCE(SUM(input_tokens),0) as total_input_tokens, COALESCE(SUM(output_tokens),0) as total_output_tokens, COALESCE(SUM(total_tokens),0) as total_tokens, COALESCE(SUM(cost),0) as total_cost, COALESCE(AVG(latency_ms),0) as avg_latency_ms").Scan(&summary)
	return &summary, nil
}

func (s *UsageService) GetTrend(userID string, query *model.UsageQuery) ([]map[string]interface{}, error) {
	startTime, _ := time.Parse("2006-01-02", query.StartDate)
	endTime, _ := time.Parse("2006-01-02", query.EndDate)
	endTime = endTime.Add(24 * time.Hour)

	tx := s.db.Model(&model.UsageRecord{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startTime, endTime)
	if query.ModelID != "" {
		tx = tx.Where("model_catalog_id = ?", query.ModelID)
	}
	if query.APIKeyID != "" {
		tx = tx.Where("api_key_id = ?", query.APIKeyID)
	}

	var results []map[string]interface{}
	tx.Select("DATE(created_at) as date, COUNT(*) as requests, SUM(total_tokens) as tokens, SUM(cost) as cost, AVG(latency_ms) as avg_latency").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results)
	return results, nil
}

func (s *UsageService) GetRecords(userID string, query *model.UsageQuery, page, pageSize int) ([]model.UsageRecord, int64, error) {
	startTime, _ := time.Parse("2006-01-02", query.StartDate)
	endTime, _ := time.Parse("2006-01-02", query.EndDate)
	endTime = endTime.Add(24 * time.Hour)

	tx := s.db.Model(&model.UsageRecord{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startTime, endTime)
	if query.ModelID != "" {
		tx = tx.Where("model_catalog_id = ?", query.ModelID)
	}
	if query.APIKeyID != "" {
		tx = tx.Where("api_key_id = ?", query.APIKeyID)
	}

	var total int64
	tx.Count(&total)

	var records []model.UsageRecord
	offset := (page - 1) * pageSize
	if err := tx.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (s *UsageService) GetAdminSummary(query *model.UsageQuery) (*model.UsageSummary, error) {
	startTime, _ := time.Parse("2006-01-02", query.StartDate)
	endTime, _ := time.Parse("2006-01-02", query.EndDate)
	endTime = endTime.Add(24 * time.Hour)

	tx := s.db.Model(&model.UsageRecord{}).Where("created_at >= ? AND created_at < ?", startTime, endTime)
	if query.ModelID != "" {
		tx = tx.Where("model_catalog_id = ?", query.ModelID)
	}

	var summary model.UsageSummary
	tx.Select("COUNT(*) as total_requests, COALESCE(SUM(input_tokens),0) as total_input_tokens, COALESCE(SUM(output_tokens),0) as total_output_tokens, COALESCE(SUM(total_tokens),0) as total_tokens, COALESCE(SUM(cost),0) as total_cost, COALESCE(AVG(latency_ms),0) as avg_latency_ms").Scan(&summary)
	return &summary, nil
}
