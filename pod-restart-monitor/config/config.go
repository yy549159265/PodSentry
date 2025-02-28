package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KubeconfigPath string
	Namespaces     []string
	TimeWindow     time.Duration
	Threshold      int
	NotifyType     string
	Webhook        string
	Rollback       bool
}

func LoadConfig() *Config {
	namespace := os.Getenv("MONITOR_NAMESPACE")
	kubeconfig := os.Getenv("KUBECONFIG_PATH")
	timeWindow := os.Getenv("TIME_WINDOW")
	threshold := os.Getenv("THRESHOLD")
	notifyType := os.Getenv("NOTIFY_TYPE")
	webhook := os.Getenv("WEBHOOK")
	rollback := os.Getenv("ROLLBACK")

	return &Config{
		KubeconfigPath: parseKubeconfig(kubeconfig),
		Namespaces:     parseNamespaces(namespace),
		TimeWindow:     parseTimeWindow(timeWindow),
		Threshold:      parseThreshold(threshold),
		NotifyType:     parseNotifyType(notifyType),
		Webhook:        parseWebhook(notifyType, webhook),
		Rollback:       parseRollback(rollback),
	}
}

func parseKubeconfig(input string) string {
	cleand := strings.TrimSpace(input)
	if cleand == "" {
		return "/app-config/kubeconfig.yaml"
	} else {
		return cleand
	}
}

func parseNamespaces(input string) []string {
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		return []string{metav1.NamespaceAll}
	}

	parts := strings.Split(cleaned, ",")
	var result []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseThreshold(input string) int {
	if input == "" {
		return 3
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		return 3
	}

	return value
}

func parseTimeWindow(input string) time.Duration {
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		cleaned = "5m"
	}

	duration, err := time.ParseDuration(cleaned)
	if err != nil {
		return 5 * time.Minute
	}
	return duration
}

func parseNotifyType(input string) string {
	if input == "" {
		return "undefined"
	}

	return strings.TrimSpace(input)
}

func parseWebhook(notifyType string, input string) string {
	if notifyType == "wechat" || notifyType == "lark" {
		return strings.TrimSpace(input)
	} else {
		return ""
	}
}

func parseRollback(input string) bool {
	if input == "" {
		return false
	}

	value, err := (strconv.ParseBool(strings.ToLower(input)))
	if err != nil {
		return false
	}

	return value
}
