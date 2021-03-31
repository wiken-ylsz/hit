package pool

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrTooBusy 当前队列任务较多, 请稍后重试
	ErrTooBusy = errors.New("当前任务较多, 请稍后重试")
	// ErrClosed 协程池已关闭
	ErrClosed = errors.New("协程池已关闭")
)
var defaultPool = NewPool(100, 10000)

// Handle 接收的任务对象
type Handle func()

// Pool 协程池
// 设计目标:
// 1.控制最大并发数量
// 2.任务高峰期增加协程执行任务, 任务空闲期关闭协程释放资源
//
// 协程的生命周期: 新建 --> 工作 --> 空闲 --> 释放
// TODO: 修复"任务高峰期后, 少量任务依然会缓存大量协程, 无任务时才会释放协程"
type Pool struct {
	limitChan int32      // 最大并发数量
	used      int32      // 已开启的协程数量
	m         sync.Mutex // 守护used变量

	free       chan chan *task // 空闲的协程
	queue      chan *task      // 任务队列
	limitQueue int32           // 队列缓存最大数量

	timeout time.Duration // 空闲协程自动清理时间

	ctx    context.Context // 同步信号
	cancel context.CancelFunc
}

// NewPool 获取一个协程池对象
func NewPool(limitChan, limitQueue int32) *Pool {
	p := new(Pool)
	p.limitChan = limitChan
	p.free = make(chan chan *task, limitChan)
	p.queue = make(chan *task, limitQueue)
	p.timeout = 10 * time.Second
	p.ctx, p.cancel = context.WithCancel(context.Background())
	go p.Run()
	return p
}

// Run 启动协程池
func (p *Pool) Run() {
	for _task := range p.queue {
		// 尝试使用空闲的协程
		select {
		case ch, ok := <-p.free:
			if !ok {
				return
			}
			ch <- _task
			continue
		case <-p.ctx.Done():
			return
		default:
		}

		// BUG: 释放的协程还在工作?
		// 尝试新增协程处理
		p.m.Lock()
		if p.used < p.limitChan {
			p.used++
			ch := make(chan *task)
			go func(ctx context.Context, ch chan *task) {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				for {
					select {
					case _task, ok := <-ch:
						if !ok {
							return
						}
						_task.do()
						_task.p.recycle(ch)
					case <-ctx.Done():
						return
					case <-time.After(p.timeout):
						_task.p.release()
					}
				}

			}(p.ctx, ch)
			ch <- _task

			p.m.Unlock()
			continue

		}
		p.m.Unlock()

		// 等待空闲的协程处理
		select {
		case ch, ok := <-p.free:
			if !ok {
				return
			}
			ch <- _task
		case <-p.ctx.Done():
			return
		}
	}
}

// Push 向队列添加新的任务
func (p *Pool) Push(h Handle) error {
	_task := newTask(p, h)
	select {
	case <-p.ctx.Done():
		return ErrClosed
	case p.queue <- _task:
	case <-time.After(p.timeout):
		return ErrTooBusy
	}

	return nil
}

// Close 发送关闭信号，释放资源
// TODO: 确保队列中的任务执行完毕后关闭
func (p *Pool) Close() {
	p.cancel()
}

// recycle 回收空闲的协程
func (p *Pool) recycle(ch chan *task) {
	select {
	case <-p.ctx.Done():
		return
	case p.free <- ch:
	default:
	}
}

// release 释放空闲的协程
func (p *Pool) release() {
	select {
	case <-p.ctx.Done():
		return
	case ch := <-p.free:
		close(ch)
		p.m.Lock()
		p.used--
		p.m.Unlock()
	default:
	}
}

type task struct {
	p *Pool
	h Handle
}

func (t *task) do() {
	t.h()
}
func newTask(p *Pool, h Handle) *task {
	return &task{
		p: p,
		h: h,
	}
}

// Push 向队列添加新的任务
func Push(h Handle) error {
	return defaultPool.Push(h)
}
