package rotate

import (
	"time"
)

// 最小批次大小
const minBatchSize = 1

// 最小自动旋转时间
const minAutoRotateTime = time.Second

// 旋转器
type Rotator interface {
	// 添加
	Add(a interface{})
	// 立即旋转
	Rotate()
}

type rotate struct {
	batchSize int           // 批次大小
	batch     []interface{} // 批次数据
	offset    int           // 批次位置偏移

	channel        chan interface{} // 通道, 用于并发转单线程
	rotateSignal   chan struct{}    // 旋转信号, 用于转单线程
	autoRotateTime time.Duration    // 自动旋转时间

	callback func(values []interface{}) // 回调
}

func (r *rotate) Add(a interface{}) {
	r.channel <- a
}

func (r *rotate) Rotate() {
	r.rotateSignal <- struct{}{}
}

// 开始
func (r *rotate) start() {
	for {
		select {
		case <-r.rotateSignal:
			r.startRotate()
		case a := <-r.channel:
			r.add(a)
		}
	}
}

// 开始自动旋转
func (r *rotate) startAutoRotate() {
	if r.autoRotateTime <= 0 {
		return
	}

	t := time.NewTicker(r.autoRotateTime)
	for {
		select {
		case <-t.C:
			r.Rotate()
		}
	}
}

// 添加数据, 这个函数必须是单线程运行的
func (r *rotate) add(a interface{}) {
	r.batch[r.offset] = a
	r.offset++

	// 检查批次
	if r.offset == r.batchSize {
		r.startRotate()
	}
}

// 开始旋转, 这个函数必须是单线程运行的
func (r *rotate) startRotate() {
	if r.offset == 0 {
		return
	}

	// 获取数据
	data := r.batch[:r.offset]

	// 重设数据
	r.batch = make([]interface{}, r.batchSize)
	r.offset = 0

	// 回调
	go r.callback(data)
}

func (r *rotate) Apply(opts []Option) {
	for _, o := range opts {
		o(r)
	}
	if r.batchSize < minBatchSize {
		r.batchSize = minBatchSize
	}
	if r.autoRotateTime > 0 && r.autoRotateTime < minAutoRotateTime {
		r.autoRotateTime = minAutoRotateTime
	}
}

func NewRotate(callback func(values []interface{}), opts ...Option) Rotator {
	r := &rotate{
		channel:      make(chan interface{}),
		rotateSignal: make(chan struct{}),
		callback:     callback,
	}
	r.Apply(opts)
	r.batch = make([]interface{}, r.batchSize)

	go r.start()
	go r.startAutoRotate()
	return r
}
