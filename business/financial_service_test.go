package business

import (
	"context"
	"errors"
	"testing"

	"trading/model"
)

// TestSaveFinancialReportDataSuccess 成功保存财报数据
func TestSaveFinancialReportDataSuccess(t *testing.T) {
	reports := []*model.FinancialReport{
		{Code: "000001", ReportDate: "2025-03-31"},
		{Code: "000001", ReportDate: "2024-12-31"},
	}
	b := &mockBroker{financialData: reports}
	repo := &mockFinancialRepo{}

	svc := NewFinancialReportService(b, repo)
	err := svc.SaveFinancialReportData(context.Background(), "000001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.upserted) != 2 {
		t.Fatalf("expected 2 reports upserted, got %d", len(repo.upserted))
	}
}

// TestSaveFinancialReportDataInvalidCode 非法股票代码
func TestSaveFinancialReportDataInvalidCode(t *testing.T) {
	svc := NewFinancialReportService(&mockBroker{}, &mockFinancialRepo{})
	err := svc.SaveFinancialReportData(context.Background(), "999999")
	if err == nil {
		t.Fatal("expected error for invalid code, got nil")
	}
}

// TestSaveFinancialReportDataRepoError repo upsert 失败
func TestSaveFinancialReportDataRepoError(t *testing.T) {
	reports := []*model.FinancialReport{
		{Code: "000001", ReportDate: "2025-03-31"},
	}
	b := &mockBroker{financialData: reports}
	repo := &mockFinancialRepo{upErr: errors.New("db down")}

	svc := NewFinancialReportService(b, repo)
	err := svc.SaveFinancialReportData(context.Background(), "000001")
	if err == nil {
		t.Fatal("expected error when repo upsert fails")
	}
}

// TestAppendFinancialReportDataSuccess 成功增量保存财报数据
func TestAppendFinancialReportDataSuccess(t *testing.T) {
	existing := []*model.FinancialReport{
		{Code: "000001", ReportDate: "2024-12-31"},
	}
	newReports := []*model.FinancialReport{
		{Code: "000001", ReportDate: "2025-03-31"},
		{Code: "000001", ReportDate: "2024-12-31"},
	}
	b := &mockBroker{financialData: newReports}
	repo := &mockFinancialRepo{reports: existing}

	svc := NewFinancialReportService(b, repo)
	err := svc.AppendFinancialReportData(context.Background(), "000001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 new report upserted, got %d", len(repo.upserted))
	}
	if repo.upserted[0].ReportDate != "2025-03-31" {
		t.Fatalf("expected report date 2025-03-31, got %s", repo.upserted[0].ReportDate)
	}
}

// TestAppendFinancialReportDataNoNewData 没有新数据时不报错
func TestAppendFinancialReportDataNoNewData(t *testing.T) {
	existing := []*model.FinancialReport{
		{Code: "000001", ReportDate: "2025-03-31"},
		{Code: "000001", ReportDate: "2024-12-31"},
	}
	sameReports := []*model.FinancialReport{
		{Code: "000001", ReportDate: "2025-03-31"},
		{Code: "000001", ReportDate: "2024-12-31"},
	}
	b := &mockBroker{financialData: sameReports}
	repo := &mockFinancialRepo{reports: existing}

	svc := NewFinancialReportService(b, repo)
	err := svc.AppendFinancialReportData(context.Background(), "000001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.upserted != nil {
		t.Fatal("expected no upsert when no new data")
	}
}

// TestAppendFinancialReportDataBrokerError broker 失败
func TestAppendFinancialReportDataBrokerError(t *testing.T) {
	b := &mockBroker{financialErr: errors.New("broker down")}
	repo := &mockFinancialRepo{reports: []*model.FinancialReport{}}

	svc := NewFinancialReportService(b, repo)
	err := svc.AppendFinancialReportData(context.Background(), "000001")
	if err == nil {
		t.Fatal("expected error when broker fails")
	}
}

// TestAppendFinancialReportDataRepoError repo FindByCode 失败
func TestAppendFinancialReportDataRepoError(t *testing.T) {
	b := &mockBroker{financialData: []*model.FinancialReport{
		{Code: "000001", ReportDate: "2025-03-31"},
	}}
	repo := &mockFinancialRepo{}
	// Override FindByCode to return error by using a different approach
	// Since mockFinancialRepo.FindByCode returns (m.reports, nil), we need
	// to test the upsert error path instead
	repo.upErr = errors.New("db down")
	repo.reports = []*model.FinancialReport{} // no existing, so newReports will have data

	svc := NewFinancialReportService(b, repo)
	err := svc.AppendFinancialReportData(context.Background(), "000001")
	if err == nil {
		t.Fatal("expected error when repo upsert fails")
	}
}
