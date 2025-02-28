package main

import (
	"context"
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/internal/k8sclient"
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/pod-restart-monitor/config"
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/pod-restart-monitor/monitor"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	clientset, err := k8sclient.NewClient(cfg.KubeconfigPath)
	if err != nil {
		logrus.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	watcher := monitor.NewPodWatcher(clientset, cfg)

	// 优雅退出处理
	// 创建一个带有信号通知的 Context
	// 当进程收到这些信号时，Go 的 signal 包会自动调用 cancel() 函数
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,  // Ctrl+C 信号
		syscall.SIGTERM, // kill 命令默认信号
	)
	defer cancel()

	for _, ns := range cfg.Namespaces {
		go func(namespace string) {
			for {
				watch, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{})
				if err != nil {
					logrus.Error("Error creating watcher:", err)
					time.Sleep(5 * time.Second)
					continue
				}

				for event := range watch.ResultChan() {
					pod, ok := event.Object.(*v1.Pod)
					if !ok {
						continue
					}
					if event.Type == "MODIFIED" {
						monitor.HandlePodEvent(watcher, pod)
					}
				}
			}
		}(ns)
	}

	// 启动清理协程
	go monitor.StartCleanupRoutine(ctx, watcher, cfg)

	// 阻塞当前 Goroutine，直到调用cancel()，才会继续执行
	<-ctx.Done()
	time.Sleep(10 * time.Second) // 等待资源释放
}
