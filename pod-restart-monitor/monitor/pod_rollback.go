package monitor

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sort"
	"strconv"
)

func PodRollback(pod *v1.Pod, client kubernetes.Interface) error {
	logrus.Infof("Pod %s restarted times more than threshold, ready to rollback", pod.Name)
	deploymentName, err := findDeploymentForPod(pod, client)
	if err != nil {
		return err
	}
	if deploymentName == "" {
		return fmt.Errorf("no deployment controller found for the pod")
	}

	if err := rollbackDeployment(deploymentName, pod, client); err != nil {
		return err // 错误日志已在rollbackDeployment中记录
	}

	return nil
}

// 查找Pod关联的Deployment名称
func findDeploymentForPod(pod *v1.Pod, client kubernetes.Interface) (string, error) {
	// 遍历OwnerReferences查找ReplicaSet
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "ReplicaSet" {
			// 获取ReplicaSet详细信息
			rs, err := client.AppsV1().ReplicaSets(pod.Namespace).Get(
				context.TODO(),
				ref.Name,
				metav1.GetOptions{},
			)
			if err != nil {
				return "", err
			}

			// 从ReplicaSet的OwnerReferences查找Deployment
			for _, rsRef := range rs.OwnerReferences {
				if rsRef.Kind == "Deployment" {
					return rsRef.Name, nil
				}
			}
		}
	}
	return "", nil
}

func rollbackDeployment(deploymentName string, pod *v1.Pod, client kubernetes.Interface) error {
	// 获取目标Deployment
	deploy, err := client.AppsV1().Deployments(pod.Namespace).Get(
		context.TODO(),
		deploymentName,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// 获取所有关联的ReplicaSet
	allRS, err := getAllAssociatedReplicaSets(client, deploy)
	if err != nil {
		return fmt.Errorf("failed to get replica sets: %w", err)
	}

	if len(allRS) < 2 {
		return fmt.Errorf("not enough revision history (need at least 2, got %d)", len(allRS))
	}

	// 按版本号降序排序
	sortReplicaSetsByRevision(allRS)

	// 找到要回滚的目标版本（当前版本的上一个健康版本）
	targetRS, err := findRollbackTarget(deploy, allRS)
	if err != nil {
		return err
	}

	// 执行回滚操作
	return performSafeRollback(client, deploy, targetRS)
}

func getAllAssociatedReplicaSets(client kubernetes.Interface, deploy *appsv1.Deployment) ([]appsv1.ReplicaSet, error) {
	var (
		continueToken string
		result        []appsv1.ReplicaSet
	)

	selector := labels.Set(deploy.Spec.Selector.MatchLabels).AsSelector().String()

	for {
		rsList, err := client.AppsV1().ReplicaSets(deploy.Namespace).List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector:   selector,
				Limit:           100,
				Continue:        continueToken,
				ResourceVersion: "0", // 禁用缓存
			},
		)
		if err != nil {
			return nil, fmt.Errorf("list replica sets failed: %w", err)
		}

		// 过滤直接属于当前Deployment的RS
		for _, rs := range rsList.Items {
			if isOwnedByDeployment(&rs, deploy.Name) {
				result = append(result, rs)
			}
		}

		if rsList.Continue == "" {
			break
		}
		continueToken = rsList.Continue
	}

	return result, nil
}
func isOwnedByDeployment(rs *appsv1.ReplicaSet, deployName string) bool {
	for _, ref := range rs.OwnerReferences {
		if ref.APIVersion == "apps/v1" &&
			ref.Kind == "Deployment" &&
			ref.Name == deployName {
			return true
		}
	}
	return false
}

func sortReplicaSetsByRevision(rsList []appsv1.ReplicaSet) {
	sort.Slice(rsList, func(i, j int) bool {
		return getRevision(rsList[i]) > getRevision(rsList[j])
	})
}

// 查找回滚目标版本
func findRollbackTarget(deploy *appsv1.Deployment, rsList []appsv1.ReplicaSet) (*appsv1.ReplicaSet, error) {
	currentRev, _ := strconv.Atoi(deploy.Annotations["deployment.kubernetes.io/revision"])
	if currentRev == -1 {
		return nil, fmt.Errorf("invalid current revision")
	}

	// 查找第一个健康的历史版本
	for _, rs := range rsList {
		rev := getRevision(rs)
		if rev < currentRev && isHealthy(&rs) {
			return &rs, nil
		}
	}

	return nil, fmt.Errorf("no healthy previous version available")
}
func getRevision(rs appsv1.ReplicaSet) int {
	revStr := rs.Annotations["deployment.kubernetes.io/revision"]
	if revStr == "" {
		return -1
	}
	rev, err := strconv.Atoi(revStr)
	if err != nil {
		return -1
	}
	return rev
}

// 检查RS是否健康
func isHealthy(rs *appsv1.ReplicaSet) bool {
	return rs.Status.ReadyReplicas > 0 &&
		rs.Status.Replicas == rs.Status.ReadyReplicas
}

// 安全回滚操作
func performSafeRollback(client kubernetes.Interface, deploy *appsv1.Deployment, targetRS *appsv1.ReplicaSet) error {
	// 创建更新对象
	newDeploy := deploy.DeepCopy()

	// 回滚模板配置
	newDeploy.Spec.Template = targetRS.Spec.Template
	_, err := client.AppsV1().Deployments(deploy.Namespace).Update(
		context.TODO(),
		newDeploy,
		metav1.UpdateOptions{},
	)
	return err
}
