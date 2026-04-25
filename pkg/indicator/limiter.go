package indicator

import "context"

// Limiter 基于 channel 的并发限流器
type Limiter struct {
	ch chan struct{}
}

// NewLimiter 创建限流器，maxConcurrent 为最大并发数
func NewLimiter(maxConcurrent int) *Limiter {
	return &Limiter{ch: make(chan struct{}, maxConcurrent)}
}

// Acquire 获取一个执行槽位，阻塞直到获取成功或 context 取消
func (l *Limiter) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case l.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release 释放一个执行槽位
func (l *Limiter) Release() {
	select {
	case <-l.ch:
	default:
	}
}
