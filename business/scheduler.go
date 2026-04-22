package business

import (
	"context"
	"fmt"
	"log"
	"time"

	"trading/data"
)

// Scheduler 定时任务调度器，每天扫描并补充缺失的股票数据
type Scheduler struct {
	svc        StockService
	dailyRepo  data.StockKlineDailyRepo
	weeklyRepo data.StockKlineWeeklyRepo
	ticker     *time.Ticker
	stopCh     chan struct{}
	interval   time.Duration
}

// NewScheduler 创建 Scheduler 实例
func NewScheduler(svc StockService, dailyRepo data.StockKlineDailyRepo, weeklyRepo data.StockKlineWeeklyRepo) *Scheduler {
	return &Scheduler{
		svc:        svc,
		dailyRepo:  dailyRepo,
		weeklyRepo: weeklyRepo,
		stopCh:     make(chan struct{}),
		interval:   10 * time.Second,
	}
}

// Start 启动调度器，按指定时间每天执行扫描
func (s *Scheduler) Start(ctx context.Context, hour, minute int) {
	go s.run(ctx, hour, minute)
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) run(ctx context.Context, hour, minute int) {
	for {
		nextRun := nextDailyTime(hour, minute)
		wait := time.Until(nextRun)
		log.Printf("[scheduler] next scan at %v (in %v)", nextRun.Format("2006-01-02 15:04:05"), wait)

		select {
		case <-time.After(wait):
			s.scanAndConsume(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) scanAndConsume(ctx context.Context) {
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

	for i, task := range tasks {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		log.Printf("[scheduler] [%d/%d] processing %s (daily=%v weekly=%v)", i+1, len(tasks), task.code, task.needDaily, task.needWeekly)
		if err := s.process(ctx, task); err != nil {
			log.Printf("[scheduler] [%d/%d] failed %s: %v", i+1, len(tasks), task.code, err)
		} else {
			log.Printf("[scheduler] [%d/%d] success %s", i+1, len(tasks), task.code)
		}

		if i < len(tasks)-1 {
			time.Sleep(s.interval)
		}
	}

	log.Println("[scheduler] all tasks done")
}

// scan 扫描所有股票代码，返回需要更新的任务列表
func (s *Scheduler) scan(ctx context.Context) ([]task, error) {
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

	var tasks []task
	for code := range codeSet {
		needDaily, needWeekly, err := s.checkCode(ctx, code, today, lastFriday)
		if err != nil {
			log.Printf("[scheduler] check %s failed: %v", code, err)
			continue
		}
		if needDaily || needWeekly {
			tasks = append(tasks, task{code: code, needDaily: needDaily, needWeekly: needWeekly})
		}
	}

	return tasks, nil
}

func (s *Scheduler) checkCode(ctx context.Context, code, today, lastFriday string) (bool, bool, error) {
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

func (s *Scheduler) process(ctx context.Context, t task) error {
	if t.needDaily {
		return s.svc.SaveHistoricalData(ctx, t.code)
	}
	if t.needWeekly {
		return s.svc.SaveHistoricalData(ctx, t.code)
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
