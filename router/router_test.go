package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/aigate/config"
	"github.com/aigate/model"
	"github.com/aigate/pkg/auth"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/pkg/response"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.Init("error")
}

func testDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	db.AutoMigrate(&model.Tenant{}, &model.User{}, &model.Role{}, &model.Gateway{}, &model.GatewayPolicy{}, &model.MetricRecord{})
	// 创建默认租户
	db.Create(&model.Tenant{ID: "t1", Name: "Test", Status: 1})
	// 创建系统角色
	db.Create(&model.Role{ID: "role-admin", Name: "超级管理员", TenantAccess: true, GatewayAccess: true, MonitorAccess: true, IsSystem: true})
	db.Create(&model.Role{ID: "role-user", Name: "普通用户", TenantAccess: false, GatewayAccess: true, MonitorAccess: true, IsSystem: true})
	// 创建 admin 用户 (对应 testToken 的 user_id "u1")
	db.Create(&model.User{ID: "u1", TenantID: "t1", Username: "tester", Password: "dummy", RoleID: "role-admin", Role: "admin", Status: 1})
	return db
}

func testConfig() *config.Config {
	return &config.Config{
		Server:   config.ServerConfig{Port: "8080", Mode: "test"},
		LogLevel: "error",
		Providers: map[string]config.ProviderConfig{
			"openai": {Enabled: false},
		},
	}
}

// 生成测试用 JWT Token
func testToken() string {
	token, _ := auth.GenerateToken("u1", "t1", "tester", "admin")
	return token
}

// ===== 健康检查 =====

func TestSetup(t *testing.T) {
	r := Setup(testConfig(), testDB())
	if r == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestAPI_Health_OK(t *testing.T) {
	r := Setup(testConfig(), testDB())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

// ===== 认证接口 =====

func TestAPI_Auth_Login(t *testing.T) {
	db := testDB()
	// 创建用户 (sha256 of "pass123")
	db.Create(&model.User{ID: "u-login", TenantID: "t1", Username: "test", Password: "9b8769a4a742959a2d0298c36fb70623f2dfacda8436237df08d8dfd5b37374c", RoleID: "role-user", Role: "user", Status: 1})

	r := Setup(testConfig(), db)
	body, _ := json.Marshal(map[string]string{"username": "test", "password": "pass123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
}

func TestAPI_Auth_LoginFail(t *testing.T) {
	r := Setup(testConfig(), testDB())
	body, _ := json.Marshal(map[string]string{"username": "nouser", "password": "wrong"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAPI_Auth_Register(t *testing.T) {
	r := Setup(testConfig(), testDB())
	body, _ := json.Marshal(map[string]string{"username": "newuser", "password": "abc123", "tenant_id": "t1"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

// ===== 需要认证的接口（加 Bearer Token）=====

func authReq(method, path string, body []byte) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	var req *http.Request
	if reader != nil {
		req = httptest.NewRequest(method, path, reader)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken())
	return req
}

func TestAPI_Gateways_CRUD(t *testing.T) {
	r := Setup(testConfig(), testDB())

	// 创建网关
	body, _ := json.Marshal(map[string]interface{}{
		"name": "DeepSeek GW", "provider": "deepseek", "base_url": "https://api.deepseek.com/v1", "model": "deepseek-chat",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authReq("POST", "/api/v1/gateways", body))
	if w.Code != http.StatusOK {
		t.Fatalf("create: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var createResp response.R
	json.Unmarshal(w.Body.Bytes(), &createResp)
	gwData := createResp.Data.(map[string]interface{})
	gwID := gwData["id"].(string)

	// 列表
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/gateways", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}

	// 获取单个
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/gateways/"+gwID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w.Code)
	}

	// 更新
	updateBody, _ := json.Marshal(map[string]interface{}{"name": "Updated GW"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("PUT", "/api/v1/gateways/"+gwID, updateBody))
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// 更新策略
	policyBody, _ := json.Marshal(map[string]interface{}{
		"rate_limit_enabled": true, "rate_limit_qps": 50,
		"circuit_breaker_enabled": true, "circuit_breaker_threshold": 0.3,
	})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("PUT", "/api/v1/gateways/"+gwID+"/policy", policyBody))
	if w.Code != http.StatusOK {
		t.Fatalf("policy: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// 删除
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("DELETE", "/api/v1/gateways/"+gwID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", w.Code)
	}
}

// ===== 监控接口 =====

func TestAPI_Metrics(t *testing.T) {
	db := testDB()
	// 插入测试数据
	db.Create(&model.MetricRecord{TenantID: "t1", GatewayID: "gw1", RequestCount: 100, TokensUsed: 5000, AvgLatencyMs: 200, ErrorCount: 5, UniqueUsers: 10})

	r := Setup(testConfig(), db)

	// 汇总
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/metrics/summary?start_date=2020-01-01&end_date=2030-12-31", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("summary: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// 趋势
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/metrics/trend?start_date=2020-01-01&end_date=2030-12-31", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("trend: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

// ===== 未认证访问被拒 =====

func TestAPI_Unauthorized(t *testing.T) {
	r := Setup(testConfig(), testDB())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/gateways", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ===== CORS =====

func TestAPI_CORS_Preflight(t *testing.T) {
	r := Setup(testConfig(), testDB())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/v1/gateways", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header")
	}
}

// ===== 管理员租户管理 =====

func TestAPI_Admin_Tenants_CRUD(t *testing.T) {
	db := testDB()
	r := Setup(testConfig(), db)

	// 创建租户
	body, _ := json.Marshal(map[string]string{"name": "NewTenant", "admin_user": "newtenant_admin", "password": "pass123", "email": "test@example.com", "phone": "13800138000"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authReq("POST", "/api/v1/admin/tenants", body))
	if w.Code != http.StatusOK {
		t.Fatalf("create: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var createResp response.R
	json.Unmarshal(w.Body.Bytes(), &createResp)
	tData := createResp.Data.(map[string]interface{})
	tenantID := tData["id"].(string)

	// 列表
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/admin/tenants", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}

	// 获取单个
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/admin/tenants/"+tenantID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w.Code)
	}

	// 更新
	updateBody, _ := json.Marshal(map[string]interface{}{"name": "UpdatedTenant", "status": 0})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("PUT", "/api/v1/admin/tenants/"+tenantID, updateBody))
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// 使用详情
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("GET", "/api/v1/admin/tenants/"+tenantID+"/usage", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("usage: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// 删除（先移除关联用户，再删除租户）
	db.Where("tenant_id = ?", tenantID).Delete(&model.User{})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, authReq("DELETE", "/api/v1/admin/tenants/"+tenantID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

// ===== 非管理员访问管理接口被拒 =====

func TestAPI_Admin_ForbiddenForUser(t *testing.T) {
	db := testDB()
	// 创建普通用户（无 tenant_access 权限）
	db.Create(&model.User{ID: "u2", TenantID: "t1", Username: "normaluser", Password: "dummy", RoleID: "role-user", Role: "user", Status: 1})

	r := Setup(testConfig(), db)

	// 生成普通用户 Token
	userToken, _ := auth.GenerateToken("u2", "t1", "normaluser", "user")

	req := httptest.NewRequest("GET", "/api/v1/admin/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for normal user, got %d, body: %s", w.Code, w.Body.String())
	}
}

// ===== 创建重名租户被拒 =====

func TestAPI_Admin_Tenant_DuplicateName(t *testing.T) {
	r := Setup(testConfig(), testDB())

	// 第一次创建（"Test"已在testDB中存在）
	body, _ := json.Marshal(map[string]string{"name": "Test", "password": "pass123"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authReq("POST", "/api/v1/admin/tenants", body))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for duplicate, got %d, body: %s", w.Code, w.Body.String())
	}
}
