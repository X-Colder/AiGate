// Package service 监控指标服务：数据记录与趋势查询
package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/aigate/model"
)

// MetricService 监控指标业务服务
type MetricService struct {
	db *gorm.DB
}

// NewMetricService 创建监控服务
func NewMetricService(db *gorm.DB) *MetricService {
	return &MetricService{db: db}
}

// Record 记录一条监控指标数据
func (s *MetricService) Record(record *model.MetricRecord) error {
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	return s.db.Create(record).Error
}

// GetSummary 获取指定时间段的汇总数据
// tenantID 为空表示查所有（admin），非空则过滤指定租户
func (s *MetricService) GetSummary(tenantID string, query *model.MetricQuery) (*model.MetricSummary, error) {
	startTime, endTime, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	db := s.db.Model(&model.MetricRecord{}).
		Where("timestamp >= ? AND timestamp < ?", startTime, endTime)

	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if query.TenantID != "" {
		db = db.Where("tenant_id = ?", query.TenantID)
	}
	if query.GatewayID != "" {
		db = db.Where("gateway_id = ?", query.GatewayID)
	}

	var summary model.MetricSummary
	err = db.Select(`
		COALESCE(SUM(request_count), 0) as total_requests,
		COALESCE(SUM(tokens_used), 0) as total_tokens,
		COALESCE(AVG(avg_latency_ms), 0) as avg_latency_ms,
		COALESCE(SUM(error_count), 0) as total_errors,
		COALESCE(SUM(unique_users), 0) as unique_users
	`).Scan(&summary).Error

	if err != nil {
		return nil, fmt.Errorf("query summary error: %w", err)
	}

	if summary.TotalRequests > 0 {
		summary.ErrorRate = float64(summary.TotalErrors) / float64(summary.TotalRequests)
	}

	return &summary, nil
}

// GetTrend 获取趋势数据（按日聚合），用于前端折线图展示
// tenantID 为空表示查所有（admin），非空则过滤指定租户
func (s *MetricService) GetTrend(tenantID string, query *model.MetricQuery) ([]map[string]interface{}, error) {
	startTime, endTime, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	db := s.db.Model(&model.MetricRecord{}).
		Where("timestamp >= ? AND timestamp < ?", startTime, endTime)

	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if query.TenantID != "" {
		db = db.Where("tenant_id = ?", query.TenantID)
	}
	if query.GatewayID != "" {
		db = db.Where("gateway_id = ?", query.GatewayID)
	}

	var results []struct {
		Date       string  `json:"date"`
		Requests   int64   `json:"requests"`
		Tokens     int64   `json:"tokens"`
		AvgLatency float64 `json:"avg_latency"`
		Errors     int64   `json:"errors"`
		Users      int64   `json:"users"`
	}

	err = db.Select(`
		DATE(timestamp) as date,
		COALESCE(SUM(request_count), 0) as requests,
		COALESCE(SUM(tokens_used), 0) as tokens,
		COALESCE(AVG(avg_latency_ms), 0) as avg_latency,
		COALESCE(SUM(error_count), 0) as errors,
		COALESCE(SUM(unique_users), 0) as users
	`).Group("DATE(timestamp)").Order("date ASC").Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("query trend error: %w", err)
	}

	trend := make([]map[string]interface{}, len(results))
	for i, r := range results {
		trend[i] = map[string]interface{}{
			"date":        r.Date,
			"requests":    r.Requests,
			"tokens":      r.Tokens,
			"avg_latency": r.AvgLatency,
			"errors":      r.Errors,
			"users":       r.Users,
		}
	}
	return trend, nil
}

// parseDateRange 解析日期范围字符串
func parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}
	// endDate 取到当天结束
	end = end.Add(24 * time.Hour)
	return start, end, nil
}
