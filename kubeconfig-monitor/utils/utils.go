package utils

import (
	"PodSentry/kubeconfig-monitor/config"
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
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
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("Failed to create GET request")
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+cfg.RancherToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Del("User-Agent")

	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("Failed to execute GET request")
		return nil, err
	}

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Get url: %s failed, Reason: %s", url, resp.Status)
		logrus.WithFields(logrus.Fields{
			"url":    url,
			"status": resp.Status,
		}).WithError(err).Error("GET request failed")
		return nil, err
	}

	return resp, nil
}

func Post(url string, cfg *config.Config) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("Failed to create POST request")
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+cfg.RancherToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Del("User-Agent")

	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).WithError(err).Error("Failed to execute POST request")
		return nil, err
	}

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Post url: %s failed, Reason: %s", url, resp.Status)
		logrus.WithFields(logrus.Fields{
			"url":    url,
			"status": resp.Status,
		}).WithError(err).Error("POST request failed")
		return nil, err
	}

	return resp, nil
}
