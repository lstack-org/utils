package gorun

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

//UntilWithTimeout 每period时间间隔，执行一次f，直到timeout或内部调用until.Cancel或ErrorBreak
func UntilWithTimeout(f func(until Until), period, timeout time.Duration) (interface{}, error) {
	ctxWithTimeout, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	until := &defaultUntil{
		ctx:        ctxWithTimeout,
		cancelFunc: cancelFunc,
	}
	wait.UntilWithContext(ctxWithTimeout, func(i context.Context) {
		f(until)
	}, period)
	return until.Item(), until.Error()
}

//UntilWithCancel 每period时间间隔，执行一次f，直到内部调用until.Cancel或ErrorBreak
func UntilWithCancel(f func(until Until), period time.Duration) (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	until := &defaultUntil{
		ctx:        ctx,
		cancelFunc: cancel,
	}
	wait.UntilWithContext(ctx, func(i context.Context) {
		f(until)
	}, period)
	return until.Item(), until.Error()
}

//Until 用于控制循环任务的控制器
type Until interface {
	ItemSave(item interface{})
	Item() interface{}

	ErrorSave(err error)
	Error() error

	Cancel()
	Retry(err error, maxRetryTime int)
	ErrorBreak(err error)
	Ctx() context.Context
}

var _ Until = &defaultUntil{}

type defaultUntil struct {
	mutex sync.Mutex

	retryTime               int
	item                    interface{}
	latestError, breakError error
	ctx                     context.Context
	cancelFunc              context.CancelFunc
}

// Ctx 返回until中使用的上下文ctx
func (d *defaultUntil) Ctx() context.Context {
	return d.ctx
}

//Retry 当期待的条件不满足时，可以继续执行该任务，
// 并循环重试，直到定时任务超时， 或重试次数达到上限
func (d *defaultUntil) Retry(err error, maxRetryTime int) {
	if d.retryTime < maxRetryTime-1 {
		d.retryTime++
	} else {
		d.ErrorBreak(err)
	}
}

//ErrorBreak 退出定时任务，并保存一个错误
//保存的错误可以通过Error方法获取
func (d *defaultUntil) ErrorBreak(err error) {
	d.breakError = err
	d.Cancel()
}

//Item 当定时任务结束后，
// 可以通过该方法获取在定时任务过程中保存的一个对象
func (d *defaultUntil) Item() interface{} {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.item
}

//ItemSave 在定时任务过程中保存一个对象，该对象可以在定时任务结束后，
//通过Item方法获取
func (d *defaultUntil) ItemSave(item interface{}) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.item = item
}

//ErrorSave 在定时任务过程中，如果认为一个错误无需退出定时任务，则可以保存该错误，继续执行
//该方法可以重复调用，保存的错误会被重复覆盖，在定时任务超时后，可以通过Error方法获取到最新保存的错误
func (d *defaultUntil) ErrorSave(err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.latestError = err
}

//Cancel 终止定时任务
func (d *defaultUntil) Cancel() {
	if d.cancelFunc != nil {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.cancelFunc()
	}
}

//Error 获取定时任务执行结束后，产生的错误
func (d *defaultUntil) Error() error {
	if d.ctx != nil {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		err := d.ctx.Err()
		if err != nil {
			switch err.Error() {
			case "context canceled":
				return d.breakError
			default: //context deadline exceeded
				if d.latestError != nil {
					return fmt.Errorf("timeout: %v", d.latestError)
				}
				return fmt.Errorf("timeout")
			}
		}
	}
	return nil
}
