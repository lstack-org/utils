package stream

import (
	"reflect"
	"sort"
	"sync"
)

// Stream 用于处理go中数组或者切片的一个接口定义
// 功能参照java中的Stream实现
type Stream interface {
	// Filter 过滤元素
	Filter(predicate Predicate) Stream
	// Map 转换元素类型
	Map(function Function) Stream
	// FlatMap 数组扁平化
	FlatMap(function Function) Stream
	// ForEach 遍历元素并结束stream
	ForEach(consumer Consumer)
	// Peek 遍历元素并继续stream
	Peek(consumer Consumer) Stream
	// Limit 截取所有元素的前maxSize个元素
	Limit(maxSize int) Stream
	// Skip 跳过所有元素的前n个元素
	Skip(n int) Stream
	// Sorted 元素排序
	Sorted(comparator Comparator) Stream
	// Distinct 元素去重
	Distinct(comparator Comparator) Stream
	// AllMatch 返回是否所有元素都满足断言
	AllMatch(predicate Predicate) bool
	// AnyMatch 返回是否存在任意一个元素满足断言
	AnyMatch(predicate Predicate) bool
	// NoneMatch 返回是否没有一个元素满足断言
	NoneMatch(predicate Predicate) bool
	// Count 返回元素个数
	Count() int
	// Reduce 元素合并
	Reduce(function BiFunction) interface{}
	// ToSlice 元素输出到targetSlice
	ToSlice(targetSlice interface{})
	// MaxMin 获取元素中的最大值或最小值
	MaxMin(comparator Comparator) interface{}
	// FindFirst 从所有元素中找到第一个匹配的元素并返回
	FindFirst(predicate Predicate) interface{}
	// Group 元素分组，返回map，k --> 分组的键，v --> 元素数组
	Group(function Function) map[interface{}][]interface{}
}

type TerminalOp interface {
	EvaluateParallel(sourceStage *pipeline)
	EvaluateSequential(sourceStage *pipeline)
}

// Predicate 断言函数，用于判断数据元素 v 是否满足条件
// 满足时，返回true，否则false
type Predicate func(v interface{}) bool

// Function 类型转换函数，用于将元素 v 转换成其他类型并返回
type Function func(v interface{}) interface{}

// Consumer 消费型函数，用于处理元素 v，无返回值
type Consumer func(v interface{})

// Comparator 元素比较函数
type Comparator func(i, j interface{}) bool

// BiFunction 数据合并函数，将元素t ， u合并处理后，返回新元素，
// 参照reduce中的使用
type BiFunction func(t, u interface{}) interface{}

type sortData struct {
	data       []interface{}
	comparator Comparator
}

func (s *sortData) Len() int {
	return len(s.data)
}
func (s *sortData) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s *sortData) Less(i, j int) bool {
	return s.comparator(s.data[i], s.data[j])
}

type ForEachOp struct {
}

//EvaluateParallel 使用goroutine并发处理pipeline
func (f ForEachOp) EvaluateParallel(sourceStage *pipeline) {
	headStage := sourceStage.nextStage
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(sourceStage.data))
	for _, v := range sourceStage.data {
		data := v
		go func() {
			defer waitGroup.Done()
			headStage.do(headStage.nextStage, data)
		}()
	}
	waitGroup.Wait()
}

// EvaluateSequential 串行处理pipeline
func (f ForEachOp) EvaluateSequential(sourceStage *pipeline) {
	headStage := sourceStage.nextStage
	for _, v := range sourceStage.data {
		headStage.do(headStage.nextStage, v)
		if sourceStage.stop {
			break
		}
	}
}

func Parallel(arr interface{}) Stream {
	return stream(arr, true)
}

func New(arr interface{}) Stream {
	return stream(arr, false)
}

func stream(arr interface{}, parallel bool) Stream {
	nilCheck(arr)
	data := make([]interface{}, 0)
	dataValue := reflect.ValueOf(&data).Elem()
	arrValue := reflect.ValueOf(arr)
	kindCheck(arrValue)
	for i := 0; i < arrValue.Len(); i++ {
		dataValue.Set(reflect.Append(dataValue, arrValue.Index(i)))
	}
	p := &pipeline{data: data, parallel: parallel}
	p.sourceStage = p
	return p
}

var _ Stream = &pipeline{}

type pipeline struct {
	lock                    sync.Mutex
	data, tmpData           []interface{}
	previousStage           *pipeline
	sourceStage             *pipeline
	nextStage               *pipeline
	parallel, entered, stop bool
	do                      func(nextStage *pipeline, v interface{})
}

func (p *pipeline) Group(function Function) map[interface{}][]interface{} {
	nilCheck(function)
	res := make(map[interface{}][]interface{})
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			out := function(v)
			if out != nil {
				if value, ok := res[out]; ok {
					value = append(value, v)
					res[out] = value
				} else {
					res[out] = []interface{}{v}
				}
			}
		},
	}
	t.evaluate(&ForEachOp{})
	return res
}

func (p *pipeline) FlatMap(function Function) Stream {
	nilCheck(function)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			out := function(v)
			if out != nil {
				data := make([]interface{}, 0)
				dataValue := reflect.ValueOf(&data).Elem()
				arrValue := reflect.ValueOf(out)
				kindCheck(arrValue)
				for i := 0; i < arrValue.Len(); i++ {
					dataValue.Set(reflect.Append(dataValue, arrValue.Index(i)))
				}
				p.tmpData = append(p.tmpData, data...)
			}
		},
	}
	t.evaluate(&ForEachOp{})
	t.data = p.tmpData
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t

}

func (p *pipeline) FindFirst(predicate Predicate) interface{} {
	nilCheck(predicate)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				match := predicate(v)
				if match {
					p.tmpData = append(p.tmpData, v)
					p.sourceStage.stop = true
				}
			}
		},
	}
	t.evaluate(&ForEachOp{})
	if p.tmpData == nil {
		return nil
	}
	return p.tmpData[0]
}

func (p *pipeline) MaxMin(comparator Comparator) interface{} {
	nilCheck(comparator)
	return p.Reduce(func(t, u interface{}) interface{} {
		if comparator(t, u) {
			return t
		}
		return u
	})
}

func (p *pipeline) ToSlice(targetSlice interface{}) {
	nilCheck(targetSlice)
	targetValue := reflect.ValueOf(targetSlice)
	if targetValue.Kind() != reflect.Ptr {
		panic("target slice must be a pointer")
	}
	kindCheck(targetValue)
	sliceValue := reflect.Indirect(targetValue)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if v != nil {
				if p.sourceStage.parallel {
					p.lock.Lock()
					defer p.lock.Unlock()
				}
				sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(v)))
			}
		},
	}
	t.evaluate(&ForEachOp{})
}

func (p *pipeline) Reduce(function BiFunction) interface{} {
	nilCheck(function)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				p.tmpData = append(p.tmpData, v)
			} else {
				res := function(p.tmpData[0], v)
				p.tmpData[0] = res
			}
		},
	}
	t.evaluate(&ForEachOp{})
	if p.tmpData == nil {
		return nil
	}
	return p.tmpData[0]
}

func (p *pipeline) Count() int {
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	return len(p.tmpData)
}

func (p *pipeline) NoneMatch(predicate Predicate) bool {
	return !p.AnyMatch(predicate)
}

func (p *pipeline) AnyMatch(predicate Predicate) bool {
	entered, stop := p.matchOps(predicate, true)
	if entered {
		return stop
	}
	return false
}

func (p *pipeline) AllMatch(predicate Predicate) bool {
	entered, stop := p.matchOps(predicate, false)
	if entered {
		return !stop
	}
	return false
}

func (p *pipeline) matchOps(predicate Predicate, flag bool) (bool, bool) {
	nilCheck(predicate)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			p.sourceStage.entered = true
			match := predicate(v)
			if !flag {
				match = !match
			}
			if match {
				p.sourceStage.stop = true
			}
		},
	}
	t.evaluate(&ForEachOp{})
	return p.sourceStage.entered, p.sourceStage.stop
}

func (p *pipeline) Distinct(comparator Comparator) Stream {
	nilCheck(comparator)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			if p.tmpData == nil {
				p.tmpData = append(p.tmpData, v)
			} else {
				flag := true
				for _, tmp := range p.tmpData {
					if comparator(tmp, v) {
						flag = false
						break
					}
				}
				if flag {
					p.tmpData = append(p.tmpData, v)
				}
			}
		},
	}
	t.evaluate(&ForEachOp{})
	t.data = p.tmpData
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Sorted(comparator Comparator) Stream {
	nilCheck(comparator)
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	s := &sortData{data: p.tmpData, comparator: comparator}
	sort.Sort(s)
	t.data = p.tmpData
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Skip(n int) Stream {
	if n < 0 {
		n = 0
	}
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	dataLen := len(p.tmpData)
	if dataLen < n {
		n = dataLen
	}
	t.data = p.tmpData[n:]
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Limit(maxSize int) Stream {
	if maxSize < 0 {
		maxSize = 0
	}
	t := p.statefulStage()
	t.evaluate(&ForEachOp{})
	dataLen := len(p.tmpData)
	if dataLen < maxSize {
		maxSize = dataLen
	}
	t.data = p.tmpData[:maxSize]
	t.parallel = p.sourceStage.parallel
	t.sourceStage = t
	return t
}

func (p *pipeline) Peek(consumer Consumer) Stream {
	nilCheck(consumer)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			consumer(v)
			nextStage.do(nextStage.nextStage, v)
		},
	}
}

func (p *pipeline) Filter(predicate Predicate) Stream {
	nilCheck(predicate)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if predicate(v) {
				nextStage.do(nextStage.nextStage, v)
			}
		},
	}
}

func (p *pipeline) ForEach(consumer Consumer) {
	nilCheck(consumer)
	t := &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			consumer(v)
		},
	}
	t.evaluate(&ForEachOp{})
}

func (p *pipeline) Map(function Function) Stream {
	nilCheck(function)
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			nextStage.do(nextStage.nextStage, function(v))
		},
	}
}

func (p *pipeline) evaluate(op TerminalOp) {
	nilCheck(op)
	for headStage := p; headStage != nil && headStage.previousStage != nil; headStage = headStage.previousStage {
		headStage.previousStage.nextStage = headStage
	}

	if p.sourceStage.parallel {
		op.EvaluateParallel(p.sourceStage)
	} else {
		op.EvaluateSequential(p.sourceStage)
	}
}

func (p *pipeline) statefulStage() *pipeline {
	return &pipeline{
		previousStage: p,
		sourceStage:   p.sourceStage,
		do: func(nextStage *pipeline, v interface{}) {
			if p.sourceStage.parallel {
				p.lock.Lock()
				defer p.lock.Unlock()
			}
			p.tmpData = append(p.tmpData, v)
		},
	}
}

func nilCheck(v interface{}) {
	if v == nil {
		panic("nil forbidden")
	}
}

func kindCheck(v reflect.Value) {
	nilCheck(v)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		panic("type must be Array or Slice")
	}
}
