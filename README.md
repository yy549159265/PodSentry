# PodSentry 部署配置说明

本文档说明 PodSentry 应用的 `deployment.yaml` 文件中使用的环境变量配置。

---

## 环境变量配置

### Init 容器：`kubeconfig-fetcher`

- ​**RANCHER_SERVER**​  
  Rancher 服务器地址，用于获取 kubeconfig 文件  
  *示例*: `https://<RANCHER_SERVER_IP>:9400`

- ​**RANCHER_TOKEN**​  
  访问 Rancher 的 API 令牌，存储于名为 `rancher-api-credentials` 的 Kubernetes Secret  
  *示例*: `token-xxxxx:<REDACTED_TOKEN_VALUE>`

- ​**POD_NAME**​  
  当前 Pod 名称（由 Kubernetes 自动注入）  
  *示例*: `<POD_NAME>`

- ​**POD_NAMESPACE**​  
  当前 Pod 所在命名空间（由 Kubernetes 自动注入）  
  *示例*: `<NAMESPACE>`

---

### 主容器：`main-app`

- ​**MONITOR_NAMESPACE**​  
  监控的目标命名空间，留空则监控所有命名空间  
  *示例*:  
  `""`（监控所有命名空间）  
  `namespace1`（单命名空间）  
  `namespace1,namespace2`（多命名空间）

- ​**KUBECONFIG_PATH**​  
  kubeconfig 文件路径  
  *示例*: `/app-config/kubeconfig/kubeconfig`

- ​**TIME_WINDOW**​  
  统计 Pod 异常重启的时间窗口（s/m/h），不填写默认5分钟  
  *示例*:  
  `""`（使用默认值）  
  `5m`（5分钟）

- ​**THRESHOLD**​  
  触发告警的重启次数阈值，不填写默认3次  
  *示例*:  
  `""`（使用默认值）  
  `3`（自定义阈值）

- ​**NOTIFY_TYPE**​  
  告警通知方式，目前支持 `wechat`/`lark`  
  *示例*: `wechat`

- ​**WEBHOOK**​  
  告警通知的 Webhook 地址  
  *示例*: `https://<WEBHOOK_URL>`

- ​**ROLLBACK**​  
  是否自动回滚到前一版本  
  *示例*: `false`

---

## 存储卷说明

- ​**kubeconfig-shared**​  
  内存共享卷，用于 init 容器与主容器之间传递 kubeconfig 文件  
  *容量限制*: 1Mi

---

## 密钥配置

- ​**rancher-api-credentials**​  
  Kubernetes Secret，存储 Rancher API 令牌  
  *密钥字段*: `token`

---

## 部署方式

通过以下命令部署 PodSentry：
```bash
kubectl apply -f deployment.yaml