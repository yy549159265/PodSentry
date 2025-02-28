package monitor

import (
	"PodSentry/kubeconfig-monitor/config"
	"PodSentry/kubeconfig-monitor/utils"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type ClusterInfo struct {
	ID   string `json:"id"`
	Spec struct {
		DisplayName string `json:"displayName"`
	} `json:"spec"`
}
type ClusterInfoList struct {
	Data []ClusterInfo `json:"data"`
}

type GenerateKubeConfigOutput struct {
	Config string `json:"config"`
}

// Step 1: 获取集群列表
func GetClusterInfoList(cfg *config.Config) ([]ClusterInfo, error) {
	url := utils.BuildURL("%s/v1/management.cattle.io.clusters?include=id&include=spec.displayName")
	resp, err := utils.Get(url, cfg)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var clusterInfoList ClusterInfoList
	if err := json.NewDecoder(resp.Body).Decode(&clusterInfoList); err != nil {
		return nil, err
	}

	return clusterInfoList.Data, nil
}

// Step 2: 检查Pod是否存在
func CheckPodExists(clusterInfo ClusterInfo, cfg *config.Config) (bool, error) {
	if clusterInfo.Spec.DisplayName == "local" {
		return false, nil
	}
	url := utils.BuildURL("%s/k8s/clusters/%s/api/v1/namespaces/%s/pods/%s", cfg.RancherServer, clusterInfo.ID, cfg.PodNamespace, cfg.PodName)
	resp, err := utils.Get(url, cfg)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// 200表示存在，404表示不存在
	return resp.StatusCode == http.StatusOK, nil
}

// Step 3: 下载kubeconfig
func DownloadKubeConfig(clusterID string, cfg *config.Config) error {
	// 注意：这个API路径可能需要调整
	url := utils.BuildURL("%s/v3/clusters/%s?action=generateKubeconfig", cfg.RancherServer, clusterID)
	resp, err := utils.Post(url, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var kubeConfigOutput GenerateKubeConfigOutput
	if err := json.NewDecoder(resp.Body).Decode(&kubeConfigOutput); err != nil {
		return err
	}

	// 创建输出文件
	outFile, err := os.Create("/data/kubeconfig.yaml")
	//outFile, err := os.Create("kubeconfig.yaml")
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 写入 kubeconfig 内容到文件
	_, err = io.WriteString(outFile, kubeConfigOutput.Config)
	return err
}
