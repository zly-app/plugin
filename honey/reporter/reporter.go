package reporter

// 上报者
type Reporter interface {
	// 上报
	Report(logs [][]byte)
}
