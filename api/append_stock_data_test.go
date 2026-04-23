package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppendStockDataSuccess(t *testing.T) {
	svc := &mockStockService{}
	r := setupTestRouter(svc, &mockScheduler{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stocks/append", nil)
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

func TestAppendStockDataError(t *testing.T) {
	svc := &mockStockService{}
	sched := &mockScheduler{triggerErr: errors.New("trigger failed")}
	r := setupTestRouter(svc, sched)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stocks/append", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestAppendStockDataAlreadyRunning(t *testing.T) {
	svc := &mockStockService{}
	sched := &mockScheduler{alreadyRunning: true}
	r := setupTestRouter(svc, sched)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stocks/append", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}
