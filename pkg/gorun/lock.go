package gorun

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Mutex interface {
	sync.Locker
	//TryLock 尝试获取锁，获取锁失败，返回false
	TryLock() bool
	//TryLockWithTimeout 在限定时间内尝试获取锁，超时后返回false
	TryLockWithTimeout(timeout time.Duration) bool
	//TryLockWithCtx 在ctx done之前尝试获取锁，ctx done后返回false
	TryLockWithCtx(ctx context.Context) bool
}

func NewMutex() Mutex {
	return new(mutex)
}

const mutexLocked = 1 << iota

//mutex 默认实现
type mutex struct {
	in sync.Mutex
}

func (m *mutex) Lock() {
	m.in.Lock()
}

func (m *mutex) Unlock() {
	m.in.Unlock()
}

func (m *mutex) TryLock() bool {
	//0表示锁未被占用，1表示被占用
	//使用系统函数cas
	//期望的值为0，写入的值为1
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.in)), 0, mutexLocked)
}

func (m *mutex) TryLockWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.TryLockWithCtx(ctx)
}

func (m *mutex) TryLockWithCtx(ctx context.Context) bool {
	return m.lockSlow(ctx)
}

//不断尝试获取锁，直到ctx done，
//或TryLock成功
//注意，ctx不要使用context.emptyCtx ,否则ctx.Done不生效
func (m *mutex) lockSlow(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			//超时
			return false
		default:
			if m.TryLock() {
				return true
			}
		}
	}
}
