package utils

import (
	"crypto/tls"
	"e.coding.byd.com/dpc/dpcyunwei/PodSentry/kubeconfig-monitor/config"
	"fmt"
	"net/http"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func BuildURL(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func Get(url string, cfg *config.Config) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+cfg.RancherToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Del("User-Agent")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Get url: %s failed, Reason: %s", url, resp.Status)
		return nil, err
	}

	return resp, nil
}

func Post(url string, cfg *config.Config) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+cfg.RancherToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Del("User-Agent")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Post url: %s failed, Reason: %s", url, resp.Status)
		return nil, err
	}

	return resp, nil
}
