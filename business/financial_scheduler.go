package business

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"trading/data"
	"trading/pkg/indicator"
)

// FinancialScheduler 财报调度器接口
type FinancialScheduler interface {
	TriggerNow(ctx context.Context) error
}

type financialScheduler struct {
	svc         StockService
	financialRepo data.FinancialReportRepo
	running     atomic.Bool
	interval    time.Duration
	limiter     *indicator.Limiter
}

// NewFinancialScheduler 创建 FinancialScheduler 实例
func NewFinancialScheduler(svc StockService, financialRepo data.FinancialReportRepo) FinancialScheduler {
	return &financialScheduler{
		svc:         svc,
		financialRepo: financialRepo,
		interval:    5 * time.Second,
		limiter:     indicator.NewLimiter(100),
	}
}

// TriggerNow 手动触发一次财报扫描（供 API 调用）
func (s *financialScheduler) TriggerNow(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		return fmt.Errorf("another task is already running")
	}

	log.Println("[financial-scheduler] manual trigger started")
	go func() {
		defer s.running.Store(false)
		if err := s.scanAndConsume(ctx); err != nil {
			log.Printf("[financial-scheduler] scan failed: %v", err)
		}
	}()

	return nil
}

func (s *financialScheduler) scanAndConsume(ctx context.Context) error {
	codes, err := s.financialRepo.FindAllCodes(ctx)
	if err != nil {
		return fmt.Errorf("find all codes failed: %w", err)
	}

	if len(codes) == 0 {
		log.Println("[financial-scheduler] no codes found")
		return nil
	}

	log.Printf("[financial-scheduler] %d codes queued, consuming one every %v", len(codes), s.interval)

	type result struct {
		code string
		err  error
	}

	results := make(chan result, len(codes))
	var wg sync.WaitGroup

	for _, code := range codes {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			if err := s.limiter.Acquire(ctx); err != nil {
				log.Printf("[financial-scheduler] limiter acquire failed for %s: %v", c, err)
				return
			}
			defer s.limiter.Release()

			if err := s.svc.AppendFinancialReportData(ctx, c); err != nil {
				results <- result{code: c, err: err}
			}
		}(code)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var failCount int
	for r := range results {
		if r.err != nil {
			failCount++
			log.Printf("[financial-scheduler] failed %s: %v", r.code, r.err)
		}
	}

	log.Printf("[financial-scheduler] all tasks done, %d failed", failCount)
	return nil
}
