package rotate

import (
	"time"
)

type Option func(r *rotate)

// 设置批次大小
func WithBatchSize(size int) Option {
	return func(r *rotate) {
		if size > 0 {
			r.batchSize = size
		}
	}
}

// 设置自动旋转时间
func WithAutoRotateTime(t time.Duration) Option {
	return func(r *rotate) {
		r.autoRotateTime = t
	}
}
