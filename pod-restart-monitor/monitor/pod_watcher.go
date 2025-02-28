package monitor

import (
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/pod-restart-monitor/config"
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/pod-restart-monitor/notify"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
	"time"
)

type PodRecord struct {
	PodName          string
	Namespace        string
	FirstDetected    time.Time
	LastRestart      time.Time
	RestartCount     int // 时间窗口内的重启次数
	RealRestartCount int // 记录的实际重启次数（如容器重启次数的最大值）
}

type PodWatcher struct {
	client    kubernetes.Interface
	config    *config.Config
	records   map[string]PodRecord
	recordsMu sync.RWMutex
}

func NewPodWatcher(client kubernetes.Interface, cfg *config.Config) *PodWatcher {
	logrus.Info("PodWatcher created")
	return &PodWatcher{
		client:  client,
		config:  cfg,
		records: make(map[string]PodRecord),
	}
}

func HandlePodEvent(w *PodWatcher, pod *v1.Pod) {
	if !isCrashLooping(pod) || !isRestartEvent(pod) {
		return
	}
	podUID := string(pod.UID)
	now := time.Now()

	w.checkRecord(pod, podUID, now)

	if record := w.getRecord(podUID); record.RestartCount >= w.config.Threshold {
		w.handleRestartThreshold(pod, podUID, now)
	}
}

func (w *PodWatcher) checkRecord(pod *v1.Pod, podUID string, now time.Time) {

	w.recordsMu.Lock()
	defer w.recordsMu.Unlock()

	record, exists := w.records[podUID]

	if !exists {
		record = w.createNewRecord(pod, now, getRealRestartCount(pod))
		logNewRecord(pod, podUID, now)
	} else {
		record = w.updateExistingRecord(record, now, getRealRestartCount(pod))
	}
	w.records[podUID] = record

}
func (w *PodWatcher) createNewRecord(pod *v1.Pod, now time.Time, currentRealRestart int) PodRecord {
	return PodRecord{
		PodName:          pod.Name,
		Namespace:        pod.Namespace,
		FirstDetected:    now,
		LastRestart:      now,
		RestartCount:     1,
		RealRestartCount: currentRealRestart,
	}
}

func (w *PodWatcher) updateExistingRecord(record PodRecord, now time.Time, currentRealRestart int) PodRecord {
	if currentRealRestart > record.RealRestartCount {
		if now.Sub(record.FirstDetected) <= w.config.TimeWindow {
			record.RestartCount++
		} else {
			record.RestartCount = 1
			record.FirstDetected = now
		}
		record.LastRestart = now
		record.RealRestartCount = currentRealRestart
	}
	return record
}

func (w *PodWatcher) resetRecord(podUID string, now time.Time, pod *v1.Pod) {
	w.recordsMu.Lock()
	defer w.recordsMu.Unlock()

	if record, exists := w.records[podUID]; exists {
		currentRealRestart := getRealRestartCount(pod)
		record.RestartCount = 0
		record.FirstDetected = now
		record.LastRestart = now
		record.RealRestartCount = currentRealRestart
		w.records[podUID] = record
	}
}
func (w *PodWatcher) handleRestartThreshold(pod *v1.Pod, podUID string, now time.Time) {
	if w.config.Rollback {
		w.rollback(pod, podUID, now)
	} else {
		w.notify(pod, podUID, now)
	}
}

func (w *PodWatcher) rollback(pod *v1.Pod, podUID string, now time.Time) {
	err := PodRollback(pod, w.client)
	message := "Pod rollback successful"
	if err != nil {
		message = fmt.Sprintf("Pod rollback failed: %v", err)
		logrus.WithError(err).Error("Rollback failed")
	} else {
		w.resetRecord(podUID, now, pod)
	}
	w.sendRollbackMessage(pod, message)
}

func (w *PodWatcher) notify(pod *v1.Pod, podUID string, now time.Time) {
	w.resetRecord(podUID, now, pod)
	w.sendRestartMessage(pod)
}

func (w *PodWatcher) sendRestartMessage(pod *v1.Pod) {
	msg := notify.GetRestartMessage(w.config, pod)
	w.sendNotification(msg)
}

func (w *PodWatcher) sendRollbackMessage(pod *v1.Pod, message string) {
	msg := notify.GetRollbackMessage(pod, message)
	w.sendNotification(msg)
}

func (w *PodWatcher) sendNotification(msg string) {
	switch w.config.NotifyType {
	case "wechat":
		notify.SendWechatWebhook(w.config, msg)
	case "lark":
		notify.SendLarkWebhook(w.config, msg)
	default:
		logrus.Warnf("Unsupported notification type: %s", w.config.NotifyType)
	}
}

func (w *PodWatcher) getRecord(podUID string) PodRecord {
	w.recordsMu.RLock()
	defer w.recordsMu.RUnlock()
	return w.records[podUID]
}

func logNewRecord(pod *v1.Pod, podUID string, now time.Time) {
	logrus.WithFields(logrus.Fields{
		"podName":       pod.Name,
		"podUID":        podUID,
		"namespace":     pod.Namespace,
		"firstDetected": now.Format("2006-01-02 15:04:05"),
	}).Info("New record created for restarting pod")
}

func isCrashLooping(pod *v1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
			return true
		}
	}
	return false
}

func isRestartEvent(pod *v1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.RestartCount > 0 {
			return true
		}
	}

	return false
}
func getRealRestartCount(pod *v1.Pod) int {
	maxRestarts := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if int(cs.RestartCount) > maxRestarts {
			maxRestarts = int(cs.RestartCount)
		}
	}
	return maxRestarts
}
