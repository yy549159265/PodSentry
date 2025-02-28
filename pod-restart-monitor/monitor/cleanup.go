package monitor

import (
	"PodSentry/pod-restart-monitor/config"
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

func StartCleanupRoutine(ctx context.Context, watcher *PodWatcher, cfg *config.Config) {
	interval := cfg.TimeWindow / 4
	if interval < time.Minute { // 最小间隔1分钟
		interval = time.Minute
	}

	go func() {
		// 创建一个定时器
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		// 清理过期记录
		logrus.Info("Starting cleanup expire pod record")
		for {
			select {
			// 每隔interval时间执行一次清理操作
			case <-ticker.C:
				watcher.cleanupRecords()
			// 当ctx.Done()被关闭时，退出循环
			case <-ctx.Done():
				return
			}
		}
	}()
}

// sub(now,第一次检测时间) > 窗口 就删除
func (w *PodWatcher) cleanupRecords() {
	logrus.Debugf("Starting cleanup records ")
	w.recordsMu.Lock()
	defer w.recordsMu.Unlock()

	now := time.Now()
	for uid, record := range w.records {
		if now.Sub(record.LastRestart) > w.config.TimeWindow {
			logrus.WithFields(logrus.Fields{
				"podName":   record.PodName,
				"namespace": record.Namespace,
				"podUID":    uid,
				"subTime":   now.Sub(record.LastRestart),
			}).Info("Pod LastRestart time exceeded TimeWindow, deleting record")
			delete(w.records, uid)
		}
	}
	logrus.Debugf("Cleanup records done")
}
