package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"trading/business"
	"trading/model"
	"trading/pkg/utils"
)

// mockStockService 模拟 StockService
type mockStockService struct {
	saveErr  error
	analysis *business.StockAnalysisData
	getErr   error
}

func (m *mockStockService) SaveHistoricalData(ctx context.Context, code string) error {
	return m.saveErr
}

func (m *mockStockService) GetStockAnalysisData(ctx context.Context, code string) (*business.StockAnalysisData, error) {
	return m.analysis, m.getErr
}

func setupTestRouter(svc business.StockService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStockHandler(svc)
	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/analysis", h.GetStockAnalysisData)
	return r
}

// TestSaveStockHistoricalDataSuccess 成功保存
func TestSaveStockHistoricalDataSuccess(t *testing.T) {
	svc := &mockStockService{}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{"code": "000001"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stocks/historical", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Fatalf("expected code 0, got %v", resp["code"])
	}
}

// TestSaveStockHistoricalDataMissingCode 缺少 code
func TestSaveStockHistoricalDataMissingCode(t *testing.T) {
	svc := &mockStockService{}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stocks/historical", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestSaveStockHistoricalDataServiceError service 失败
func TestSaveStockHistoricalDataServiceError(t *testing.T) {
	svc := &mockStockService{saveErr: errors.New("service error")}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{"code": "000001"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stocks/historical", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestGetStockAnalysisDataSuccess 成功获取分析数据
func TestGetStockAnalysisDataSuccess(t *testing.T) {
	svc := &mockStockService{
		analysis: &business.StockAnalysisData{
			Daily: []*model.StockKline{
				{Code: "000001", Date: "2025-04-21", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
			},
			Weekly:     []*model.StockKline{},
			DailyMACD:  []utils.MACDResult{},
			WeeklyMACD: []utils.MACDResult{},
			DailyKDJ:   []utils.KDJResult{},
			WeeklyKDJ:  []utils.KDJResult{},
		},
	}
	r := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stocks/analysis?code=000001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Fatalf("expected code 0, got %v", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if len(data["daily"].([]interface{})) != 1 {
		t.Fatalf("expected 1 daily item, got %d", len(data["daily"].([]interface{})))
	}
}

// TestGetStockAnalysisDataMissingCode 缺少 code
func TestGetStockAnalysisDataMissingCode(t *testing.T) {
	svc := &mockStockService{}
	r := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stocks/analysis", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestGetStockAnalysisDataServiceError service 失败
func TestGetStockAnalysisDataServiceError(t *testing.T) {
	svc := &mockStockService{getErr: errors.New("service error")}
	r := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stocks/analysis?code=000001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
