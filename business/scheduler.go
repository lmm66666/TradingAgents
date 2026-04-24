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

// Scheduler 调度器接口
type Scheduler interface {
	Start(ctx context.Context, hour, minute int)
	Stop()
	TriggerNow(ctx context.Context) error
}

// stockScheduler 定时任务调度器实现，每天扫描并补充缺失的股票数据
type stockScheduler struct {
	svc        StockService
	dailyRepo  data.StockKlineDailyRepo
	weeklyRepo data.StockKlineWeeklyRepo
	stopCh     chan struct{}
	interval   time.Duration
	running    atomic.Bool
	bgCtx      context.Context
	limiter    *indicator.Limiter
}

// NewScheduler 创建 Scheduler 实例
func NewScheduler(svc StockService, dailyRepo data.StockKlineDailyRepo, weeklyRepo data.StockKlineWeeklyRepo) Scheduler {
	return &stockScheduler{
		svc:        svc,
		dailyRepo:  dailyRepo,
		weeklyRepo: weeklyRepo,
		stopCh:     make(chan struct{}),
		interval:   5 * time.Second,
		limiter:    indicator.NewLimiter(100),
	}
}

// Start 启动调度器，按指定时间每天执行扫描
func (s *stockScheduler) Start(ctx context.Context, hour, minute int) {
	s.bgCtx = ctx
	go s.run(ctx, hour, minute)
}

// Stop 停止调度器
func (s *stockScheduler) Stop() {
	close(s.stopCh)
}

func (s *stockScheduler) run(ctx context.Context, hour, minute int) {
	for {
		nextRun := nextDailyTime(hour, minute)
		wait := time.Until(nextRun)
		log.Printf("[scheduler] next scan at %v (in %v)", nextRun.Format("2006-01-02 15:04:05"), wait)

		select {
		case <-time.After(wait):
			if isWeekday(time.Now()) {
				s.scanAndConsume(ctx)
			} else {
				log.Println("[scheduler] skipped: not a weekday")
			}
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// TriggerNow 手动触发一次扫描（供 API 调用）
func (s *stockScheduler) TriggerNow(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		return fmt.Errorf("another task is already running")
	}

	log.Println("[scheduler] manual trigger started")
	go func() {
		defer s.running.Store(false)
		s.scanAndConsume(s.bgCtx)
	}()

	return nil
}

func (s *stockScheduler) scanAndConsume(ctx context.Context) {
	tasks, err := s.scan(ctx)
	if err != nil {
		log.Printf("[scheduler] scan failed: %v", err)
		return
	}

	if len(tasks) == 0 {
		log.Println("[scheduler] no missing data, all up to date")
		return
	}

	log.Printf("[scheduler] %d tasks queued, consuming one every %v", len(tasks), s.interval)

	for i, t := range tasks {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		log.Printf("[scheduler] [%d/%d] processing %s (daily=%v weekly=%v)", i+1, len(tasks), t.code, t.needDaily, t.needWeekly)
		if err = s.process(ctx, t); err != nil {
			log.Printf("[scheduler] [%d/%d] failed %s: %v", i+1, len(tasks), t.code, err)
		} else {
			log.Printf("[scheduler] [%d/%d] success %s", i+1, len(tasks), t.code)
		}

		if i < len(tasks)-1 {
			time.Sleep(s.interval)
		}
	}

	log.Println("[scheduler] all tasks done")
}

// scan 扫描所有股票代码，返回需要更新的任务列表
func (s *stockScheduler) scan(ctx context.Context) ([]task, error) {
	dailyCodes, err := s.dailyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find daily codes failed: %w", err)
	}

	weeklyCodes, err := s.weeklyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find weekly codes failed: %w", err)
	}

	codeSet := make(map[string]struct{})
	for _, c := range dailyCodes {
		codeSet[c] = struct{}{}
	}
	for _, c := range weeklyCodes {
		codeSet[c] = struct{}{}
	}

	today := time.Now().Format("2006-01-02")
	lastFriday := lastFridayDate(time.Now())

	type result struct {
		task task
		ok   bool
	}

	results := make(chan result, len(codeSet))
	var wg sync.WaitGroup

	for code := range codeSet {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			if err := s.limiter.Acquire(ctx); err != nil {
				log.Printf("[scheduler] limiter acquire failed for %s: %v", c, err)
				return
			}
			defer s.limiter.Release()

			needDaily, needWeekly, err := s.checkCode(ctx, c, today, lastFriday)
			if err != nil {
				log.Printf("[scheduler] check %s failed: %v", c, err)
				return
			}
			if needDaily || needWeekly {
				results <- result{task: task{code: c, needDaily: needDaily, needWeekly: needWeekly}, ok: true}
			}
		}(code)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var tasks []task
	for r := range results {
		if r.ok {
			tasks = append(tasks, r.task)
		}
	}

	return tasks, nil
}

func (s *stockScheduler) checkCode(ctx context.Context, code, today, lastFriday string) (bool, bool, error) {
	var needDaily bool
	latestDaily, err := s.dailyRepo.FindLatestByCode(ctx, code)
	if err != nil {
		needDaily = true
	} else if latestDaily.Date != today {
		needDaily = true
	}

	var needWeekly bool
	latestWeekly, err := s.weeklyRepo.FindLatestByCode(ctx, code)
	if err != nil {
		needWeekly = true
	} else if latestWeekly.Date != lastFriday {
		needWeekly = true
	}

	return needDaily, needWeekly, nil
}

func (s *stockScheduler) process(ctx context.Context, t task) error {
	if t.needDaily || t.needWeekly {
		return s.svc.AppendStockData(ctx, t.code)
	}
	return nil
}

type task struct {
	code       string
	needDaily  bool
	needWeekly bool
}

// nextDailyTime 计算下一个指定时间（今天或明天）
func nextDailyTime(hour, minute int) time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// lastFridayDate 返回最近一个周五的日期字符串
func lastFridayDate(t time.Time) string {
	wd := t.Weekday()
	daysBack := int(wd - time.Friday)
	if daysBack < 0 {
		daysBack += 7
	}
	if wd == time.Friday {
		daysBack = 0
	}
	return t.Add(-time.Duration(daysBack) * 24 * time.Hour).Format("2006-01-02")
}

// isWeekday 判断是否为工作日（周一到周五）
func isWeekday(t time.Time) bool {
	wd := t.Weekday()
	return wd != time.Saturday && wd != time.Sunday
}
