package notify

import (
	"bytes"
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/pod-restart-monitor/config"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"time"
)

// 企业微信webhook通知
func SendWechatWebhook(cfg *config.Config, message string) {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	sendHTTPRequest(cfg.Webhook, payload)
}

// 飞书webhook通知
func SendLarkWebhook(cfg *config.Config, message string) {
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": message,
		},
	}

	sendHTTPRequest(cfg.Webhook, payload)
}

// 通用HTTP请求发送函数
func sendHTTPRequest(url string, payload interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("Failed to marshal payload")

	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("HTTP request failed")

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		logrus.WithFields(logrus.Fields{
			"url":        url,
			"statusCode": resp.StatusCode,
		}).WithError(err).Error("Unexpected status code")
	}
}

// pod restart template
func GetRestartMessage(cfg *config.Config, pod *v1.Pod) string {
	return fmt.Sprintf(`
		POD: %s
		NAMESPACE: %s
		RESTARTS_THRESHOLD: %d
		TIMESTAMP: %s
		MESSAGE: pod restarted times to the threshold
		`,
		pod.Name,
		pod.Namespace,
		cfg.Threshold,
		time.Now().Format("2006-01-02 15:04:05"))
}

func GetRollbackMessage(pod *v1.Pod, err string) string {
	return fmt.Sprintf(`
		POD: %s
		NAMESPACE: %s
		TIMESTAMP: %s
		MESSAGE:  %s
		`,
		pod.Name,
		pod.Namespace,
		time.Now().Format("2006-01-02 15:04:05"),
		err)
}
