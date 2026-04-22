package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"trading/business"
)

func TestGetStockAnalysisDataSuccess(t *testing.T) {
	svc := &mockStockService{
		analysis: &business.StockAnalysisData{
			Daily: []business.AnalysisItem{
				{Date: "2025-04-21", Price: 1.5, Volume: 100, J: 50, DEA: 0.05, MA5: 1.5, MA20: 1.5, MA60: 1.5},
			},
			Weekly: []business.AnalysisItem{},
		},
	}
	r := setupTestRouter(svc, &mockScheduler{})

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

func TestGetStockAnalysisDataMissingCode(t *testing.T) {
	svc := &mockStockService{}
	r := setupTestRouter(svc, &mockScheduler{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stocks/analysis", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetStockAnalysisDataServiceError(t *testing.T) {
	svc := &mockStockService{getErr: errors.New("service error")}
	r := setupTestRouter(svc, &mockScheduler{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stocks/analysis?code=000001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
