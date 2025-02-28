package main

import (
	"PodSentry/kubeconfig-monitor/config"
	"PodSentry/kubeconfig-monitor/monitor"
	"github.com/sirupsen/logrus"
)

var cfg *config.Config

func main() {
	cfg = config.LoadConfig()
	clusterInfoList, err := monitor.GetClusterInfoList(cfg)
	if err != nil {
		logrus.WithError(err).Error("Failed to get cluster info list")
		return
	}

	found := false

	for _, clusterInfo := range clusterInfoList {
		exists, err := monitor.CheckPodExists(clusterInfo, cfg)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"cluster": clusterInfo.Spec.DisplayName,
			}).WithError(err).Error("Failed to check if pod exists")
			continue
		}

		if exists {
			logrus.WithFields(logrus.Fields{
				"podName": cfg.PodName,
				"cluster": clusterInfo.Spec.DisplayName,
			}).Info("Pod exists in cluster")

			err := monitor.DownloadKubeConfig(clusterInfo.ID, cfg)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"cluster": clusterInfo.Spec.DisplayName,
				}).WithError(err).Error("Failed to download Kubeconfig")
			} else {
				logrus.WithFields(logrus.Fields{
					"cluster": clusterInfo.ID,
				}).Info("Kubeconfig downloaded for cluster")
			}

			found = true
			break // End loop after finding the pod
		}
	}

	if !found {
		logrus.WithField("podName", cfg.PodName).Info("Pod does not exist in any cluster")
	}

}
