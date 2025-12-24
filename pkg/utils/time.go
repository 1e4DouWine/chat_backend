package utils

import (
	"fmt"
	"time"
)

// FormatLatency 自适应格式化耗时，根据耗时大小自动选择合适的单位
// 小于1微秒显示纳秒(ns)
// 小于1毫秒显示微秒(μs)
// 小于1秒显示毫秒(ms)
// 大于等于1秒显示秒(s)
func FormatLatency(d time.Duration) string {
	// 转换为纳秒
	ns := d.Nanoseconds()
	
	// 小于1微秒，显示纳秒
	if ns < 1000 {
		return fmt.Sprintf("%dns", ns)
	}
	
	// 小于1毫秒，显示微秒
	if ns < 1000000 {
		return fmt.Sprintf("%.2fμs", float64(ns)/1000)
	}
	
	// 小于1秒，显示毫秒
	if ns < 1000000000 {
		return fmt.Sprintf("%.2fms", float64(ns)/1000000)
	}
	
	// 大于等于1秒，显示秒
	return fmt.Sprintf("%.2fs", float64(ns)/1000000000)
}