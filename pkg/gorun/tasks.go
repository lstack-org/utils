package gorun

import (
	"context"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

type BatchWait interface {
	Await(ctx context.Context) (BatchRes, error)
	AwaitWithTimeout(timeout time.Duration) (BatchRes, error)
}

//BatchTaskAction 批量任务函数类型定义
type BatchTaskAction func(ctx BatchContext)

//BatchContext 批量任务上下文接口
type BatchContext interface {
	context.Context
	BatchRes
}

//defaultBatchContext BatchContext默认实现
type defaultBatchContext struct {
	context.Context
	BatchRes
}

//NewBatchContext 构建BatchContext实例
func NewBatchContext(ctx context.Context) BatchContext {
	return &defaultBatchContext{
		Context:  ctx,
		BatchRes: NewBatchRes(),
	}
}

func Tasks(acts ...BatchTaskAction) BatchWait {
	return &BatchTasks{
		wg:      &sync.WaitGroup{},
		actions: acts,
	}
}

//BatchTasks BatchWait实现类
type BatchTasks struct {
	panicI  interface{}
	wg      *sync.WaitGroup
	actions []BatchTaskAction
	BatchContext
}

//Await 等待所有goroutine执行结束
func (b *BatchTasks) Await(ctx context.Context) (BatchRes, error) {
	b.Exec(ctx)
	b.wg.Wait()
	if b.panicI != nil {
		panic(b.panicI)
	}
	return b.BatchContext, b.GetMergedError()
}

//AwaitWithTimeout 等待所有goroutine执行结束
//最多等待时长:timeout
//超时返回错误
func (b *BatchTasks) AwaitWithTimeout(timeout time.Duration) (BatchRes, error) {
	withTimeoutCtx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	batchRes, err := b.Await(withTimeoutCtx)
	if timeoutErr := withTimeoutCtx.Err(); timeoutErr != nil {
		return batchRes, timeoutErr
	}
	return batchRes, err
}

//Exec 以goroutine的方式运行BatchTaskAction函数
func (b *BatchTasks) Exec(ctx context.Context) {
	b.BatchContext = NewBatchContext(ctx)
	if len(b.actions) > 0 {
		actionSignalFunc := func(ctx BatchContext, act BatchTaskAction, actionOverSignal chan bool) {
			defer func() {
				if panicI := recover(); panicI != nil {
					b.panicI = panicI
				}
				close(actionOverSignal)
			}()
			act(ctx)
		}

		for _, v := range b.actions {
			action := v
			if action == nil {
				continue
			}
			actionOverSignal := make(chan bool, 1)
			b.wg.Add(1)
			go func() {
				actionSignalFunc(b.BatchContext, action, actionOverSignal)
			}()
			go func() {
				defer b.wg.Done()
				select {
				case <-ctx.Done():
					klog.Error(ctx.Err())
					break
				case <-actionOverSignal:
					break
				}
			}()
		}
	}
}

//BatchRes 定义批量任务处理后的返回结果
type BatchRes interface {
	GetRes() []interface{}
	GetItem() interface{}
	AddItem(item interface{})
	GetMergedError() error
	AddError(err error)
}

//NewBatchRes 构建一个BatchRes实例
func NewBatchRes() BatchRes {
	return &defaultBatchRes{}
}

var _ BatchRes = &defaultBatchRes{}

type defaultBatchRes struct {
	lock    sync.RWMutex
	errs    []error
	resList []interface{}
}

//GetItem 从批量处理结果中获取一个保存的对象
func (d *defaultBatchRes) GetItem() interface{} {
	if len(d.resList) > 0 {
		d.lock.RLock()
		defer d.lock.RUnlock()
		return d.resList[0]
	}
	return nil
}

//AddItem 保存一个对象
func (d *defaultBatchRes) AddItem(item interface{}) {
	if item != nil {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.resList = append(d.resList, item)
	}
}

//AddError 保存一个错误
func (d *defaultBatchRes) AddError(err error) {
	if err != nil {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.errs = append(d.errs, err)
	}
}

//GetRes 获取批量处理的所有结果
func (d *defaultBatchRes) GetRes() []interface{} {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.resList
}

//GetMergedError 获取批量处理中保存的合并错误
func (d *defaultBatchRes) GetMergedError() error {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return errors.NewAggregate(d.errs)
}
