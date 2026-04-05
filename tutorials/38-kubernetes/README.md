# 第三十八課：Kubernetes 部署

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解基本概念（Pod/Service/Deployment），能部署簡單應用 |
| 🔴 資深工程師 | **必備**：能設計生產環境部署，包含 HPA、資源限制、Health Check |
| 🏢 DevOps/SRE | 核心技能，負責叢集管理和維運 |

## 核心資源

| 資源 | 說明 | 本課用到 |
|------|------|---------|
| **Pod** | 最小部署單位 | ✅（由 Deployment 管理） |
| **Deployment** | 管理 Pod 副本和更新 | ✅ |
| **Service** | 穩定的網路端點 | ✅ ClusterIP |
| **ConfigMap** | 非敏感設定 | ✅ |
| **Secret** | 敏感資訊 | ✅ |
| **Ingress** | 對外 HTTP 入口 | ✅ |
| **HPA** | 自動水平擴展 | ✅ |

## 使用方式

```bash
# 部署所有資源
kubectl apply -f tutorials/38-kubernetes/

# 查看 Pod 狀態
kubectl get pods -n my-app

# 查看 Deployment 狀態
kubectl rollout status deployment/my-app -n my-app

# 查看日誌
kubectl logs -f deployment/my-app -n my-app

# 進入 Pod 除錯
kubectl exec -it deployment/my-app -n my-app -- sh
```

## 重要概念

### 零停機部署（Rolling Update）

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1        # 最多額外多幾個 Pod
    maxUnavailable: 0  # 保證所有 Pod 都可用
```

### Health Check 三劍客

```yaml
livenessProbe:  # 存活探針（失敗 → 重啟 Pod）
  httpGet: {path: /healthz, port: 8080}

readinessProbe: # 就緒探針（失敗 → 從 Service 移除）
  httpGet: {path: /readyz, port: 8080}

startupProbe:   # 啟動探針（給慢啟動應用更多時間）
  failureThreshold: 30
```

### 資源管理

```yaml
resources:
  requests:  # 排程保證（Pod 一定有這麼多資源）
    cpu: "100m"
    memory: "128Mi"
  limits:    # 上限（超過 CPU 被限速，超過記憶體被殺）
    cpu: "500m"
    memory: "256Mi"
```

## 本地開發工具

```bash
# 本地 K8s 叢集
minikube start   # Minikube
kind create cluster  # Kind（Docker 內執行）

# 快速部署
helm install my-app ./chart  # Helm Chart
```
