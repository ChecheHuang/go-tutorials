# 第三十八課：Kubernetes 部署

> **一句話總結**：Kubernetes 是貨櫃碼頭的管理員——你只要說「我要 3 個 API 伺服器」，它會自動安排碼頭、搬運貨櫃、處理故障。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解基本概念（Pod/Service/Deployment），能部署簡單應用 |
| 🔴 資深工程師 | **必備**：能設計生產環境部署，包含 HPA、資源限制、Health Check |
| 🏢 DevOps/SRE | 核心技能，負責叢集管理和維運 |

## 你會學到什麼？

- Docker 和 Kubernetes 的關係：貨櫃 vs 碼頭
- K8s 核心資源：Pod、Deployment、Service、Ingress、ConfigMap、Secret
- `deployment.yaml` 逐行解讀
- Service 三種類型：ClusterIP、NodePort、LoadBalancer
- Health Check 三劍客：Liveness、Readiness、Startup Probe
- HPA 水平自動擴展設定
- Rolling Update 零停機部署與回滾
- 資源 Requests / Limits 為何重要
- kubectl 常用指令速查

## 執行方式

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

---

## 生活比喻：Kubernetes = 貨櫃碼頭管理員

```
你（開發者）         碼頭管理員（K8s）          碼頭（Node）
    │                      │                      │
    │  "我要 3 個 API"     │                      │
    ├─────────────────────→│                      │
    │                      │  安排 3 個貨櫃位     │
    │                      ├─────────────────────→│
    │                      │                      │  [Pod-1] [Pod-2] [Pod-3]
    │                      │                      │
    │                      │  貨櫃-2 壞了！        │
    │                      │  自動搬一個新的上去    │
    │                      ├─────────────────────→│
    │                      │                      │  [Pod-1] [Pod-3] [Pod-4]
```

| 現實概念 | Kubernetes 對應 | 說明 |
|---------|----------------|------|
| 貨櫃 | Container（Docker） | 裝好的應用程式 |
| 貨櫃船 | Pod | 一個或多個貨櫃一起運輸 |
| 碼頭泊位 | Node | 實際的機器（VM 或實體機） |
| 碼頭管理員 | Control Plane | 決定貨櫃放哪裡 |
| 提貨單 | Deployment | 你想要幾個、什麼版本 |
| 收貨窗口 | Service | 固定的取貨地址 |
| 海關入口 | Ingress | 外部流量的入口 |

---

## Docker vs Kubernetes

很多初學者搞混 Docker 和 Kubernetes 的定位。用貨運來比喻：

```
Docker（貨櫃）                    Kubernetes（碼頭）
┌─────────────────────┐          ┌──────────────────────────┐
│  打包你的應用程式     │          │  管理 100 個貨櫃放在哪    │
│  一個 Dockerfile     │          │  壞了自動換新的           │
│  → 一個 Image        │          │  流量自動分配             │
│  → 一個 Container    │          │  自動擴縮容               │
└─────────────────────┘          └──────────────────────────┘

Docker 解決：「在我的電腦上可以跑」→ 哪裡都能跑
K8s   解決：「一個容器可以跑」  → 一千個容器怎麼管
```

| 比較項目 | Docker | Kubernetes |
|---------|--------|------------|
| 角色 | 打包和執行容器 | 編排和管理容器 |
| 粒度 | 單一容器 | 多容器叢集 |
| 自動修復 | ❌ 容器掛了就掛了 | ✅ 自動重啟、重新排程 |
| 負載均衡 | ❌ 需要自己設 | ✅ Service 內建 |
| 自動擴展 | ❌ | ✅ HPA / VPA |
| 設定管理 | 環境變數、Volume | ConfigMap、Secret |
| 適合場景 | 開發、CI/CD | 生產環境部署 |

---

## 核心資源一覽

| 資源 | 說明 | 本課用到 |
|------|------|---------|
| **Pod** | 最小部署單位，包含一或多個容器 | ✅（由 Deployment 管理） |
| **Deployment** | 管理 Pod 副本數量和更新策略 | ✅ |
| **Service** | 穩定的網路端點，負載均衡 | ✅ ClusterIP |
| **ConfigMap** | 非敏感設定（config 檔、環境變數） | ✅ |
| **Secret** | 敏感資訊（密碼、Token、TLS 憑證） | ✅ |
| **Ingress** | 對外 HTTP/HTTPS 入口，路由規則 | ✅ |
| **HPA** | 自動水平擴展 | ✅ |
| **Namespace** | 資源隔離（類似資料夾） | ✅ |

### 資源之間的關係

```
                        外部流量
                           │
                    ┌──────▼──────┐
                    │   Ingress   │  ← 路由規則：/api → my-app-svc
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   Service   │  ← ClusterIP，內部 DNS：my-app-svc
                    └──┬───┬───┬──┘
                       │   │   │     ← 負載均衡
                ┌──────▼┐ ┌▼──────┐ ┌▼──────┐
                │ Pod-1 │ │ Pod-2 │ │ Pod-3 │  ← Deployment 管理副本數
                └───────┘ └───────┘ └───────┘
                    │         │         │
              ┌─────▼─────────▼─────────▼─────┐
              │  ConfigMap / Secret（掛載）     │
              └───────────────────────────────┘
```

---

## deployment.yaml 逐行解讀

```yaml
apiVersion: apps/v1          # API 版本，Deployment 用 apps/v1
kind: Deployment             # 資源類型
metadata:
  name: my-app               # Deployment 的名字
  namespace: my-app          # 放在哪個 namespace
  labels:
    app: my-app              # 標籤，給 Service 做選擇器用
spec:
  replicas: 3                # 要幾個 Pod（副本數）
  selector:
    matchLabels:
      app: my-app            # 要管理哪些 Pod（透過 label 匹配）
  strategy:
    type: RollingUpdate      # 更新策略：滾動更新
    rollingUpdate:
      maxSurge: 1            # 更新時最多額外多 1 個 Pod
      maxUnavailable: 0      # 保證所有舊 Pod 都可用才換新的
  template:                  # Pod 的模板
    metadata:
      labels:
        app: my-app          # Pod 的標籤（必須跟 selector 匹配）
    spec:
      containers:
      - name: api            # 容器名稱
        image: my-app:v1.2   # Docker Image
        ports:
        - containerPort: 8080  # 容器監聽的 port
        env:
        - name: DB_HOST                      # 環境變數直接寫
          value: "postgres-svc"
        - name: JWT_SECRET                   # 從 Secret 讀取
          valueFrom:
            secretKeyRef:
              name: my-app-secret
              key: jwt-secret
        envFrom:
        - configMapRef:
            name: my-app-config              # 整個 ConfigMap 匯入為環境變數
        resources:
          requests:            # 最低保證資源
            cpu: "100m"        # 0.1 核
            memory: "128Mi"    # 128 MB
          limits:              # 資源上限
            cpu: "500m"        # 0.5 核
            memory: "256Mi"    # 256 MB
        livenessProbe:         # 存活探針
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:        # 就緒探針
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 5
```

---

## Service 類型比較

Service 是 Pod 前面的穩定入口。Pod IP 會變，但 Service IP（ClusterIP）不會。

| 類型 | 存取範圍 | 用途 | 範例 |
|------|---------|------|------|
| **ClusterIP** | 叢集內部 | 服務間通訊 | API → DB、API → Redis |
| **NodePort** | 叢集外部（Node IP:Port） | 開發/測試用 | `http://node-ip:30080` |
| **LoadBalancer** | 外部（雲端 LB） | 生產環境對外 | AWS ALB / GCP LB |

```yaml
# ClusterIP（最常用）
apiVersion: v1
kind: Service
metadata:
  name: my-app-svc
  namespace: my-app
spec:
  type: ClusterIP            # 預設值，可省略
  selector:
    app: my-app              # 選擇哪些 Pod
  ports:
  - port: 80                 # Service 的 port
    targetPort: 8080         # Pod 的 port
```

```
叢集內其他 Pod 透過 DNS 存取：
  http://my-app-svc.my-app.svc.cluster.local:80
  └─ Service 名 ─┘ └ NS ─┘ └── 固定後綴 ──────┘

簡寫（同 namespace 內）：
  http://my-app-svc:80
```

---

## Health Check 三劍客

```
                              Pod 啟動
                                 │
                          ┌──────▼──────┐
                          │ startupProbe │  ← 啟動探針：給慢啟動應用時間
                          │ 最多等 300s  │     失敗 → 重啟 Pod
                          └──────┬──────┘
                                 │ 成功
                     ┌───────────┼───────────┐
               ┌─────▼─────┐          ┌─────▼──────┐
               │livenessProbe│         │readinessProbe│
               │ 每 10s 檢查 │         │ 每 5s 檢查   │
               └─────┬─────┘          └─────┬──────┘
                     │                       │
              失敗 → 重啟 Pod         失敗 → 從 Service 移除
              （Pod 死了）            （Pod 還活著但沒準備好）
```

| 探針 | 失敗後果 | 適用場景 | 檢查內容 |
|------|---------|---------|---------|
| **livenessProbe** | 重啟 Pod | 偵測死鎖、無回應 | `GET /healthz` → 200 |
| **readinessProbe** | 從 Service 移除 | DB 連線還沒建好 | `GET /readyz` → 200（含依賴檢查）|
| **startupProbe** | 重啟 Pod | 需要載入大量資料 | 同 liveness，但只跑一次 |

```yaml
# 完整設定範例
containers:
- name: api
  livenessProbe:
    httpGet:
      path: /healthz
      port: 8080
    initialDelaySeconds: 5     # 啟動後等 5 秒才開始檢查
    periodSeconds: 10          # 每 10 秒檢查一次
    timeoutSeconds: 3          # 每次檢查超時 3 秒
    failureThreshold: 3        # 連續失敗 3 次才判定
  readinessProbe:
    httpGet:
      path: /readyz
      port: 8080
    initialDelaySeconds: 3
    periodSeconds: 5
    failureThreshold: 2
  startupProbe:
    httpGet:
      path: /healthz
      port: 8080
    failureThreshold: 30       # 最多等 30 * 10 = 300 秒
    periodSeconds: 10
```

**`/healthz` vs `/readyz` 的差別**：

```go
// /healthz — 我還活著嗎？（輕量檢查）
func healthz(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
}

// /readyz — 我準備好接收流量了嗎？（含依賴檢查）
func readyz(c *gin.Context) {
    if err := db.Ping(); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "db": err.Error()})
        return
    }
    if err := rdb.Ping(ctx).Err(); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "redis": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "ready"})
}
```

---

## HPA（Horizontal Pod Autoscaler）

HPA 根據 CPU / Memory / 自訂指標自動調整 Pod 數量。

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app-hpa
  namespace: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app               # 要擴展哪個 Deployment
  minReplicas: 2               # 最少 2 個 Pod
  maxReplicas: 10              # 最多 10 個 Pod
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70  # CPU 平均使用率 > 70% 就擴展
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80  # Memory 平均使用率 > 80% 就擴展
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 30   # 擴展前等 30 秒確認
      policies:
      - type: Pods
        value: 2               # 每次最多加 2 個 Pod
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300  # 縮容前等 5 分鐘確認
      policies:
      - type: Pods
        value: 1               # 每次最多減 1 個 Pod
        periodSeconds: 60
```

> **重點**：HPA 需要 Deployment 設定 `resources.requests`，否則無法計算使用率百分比。

---

## Rolling Update 與回滾

### 零停機更新流程

```
更新前：[v1] [v1] [v1]  （3 個 Pod）

Step 1：[v1] [v1] [v1] [v2]   maxSurge=1，先加一個 v2
Step 2：[v1] [v1] [v2] [v2]   v2 就緒後，替換一個 v1
Step 3：[v1] [v2] [v2] [v2]   繼續替換
Step 4：[v2] [v2] [v2]        全部替換完成
```

### 常用指令

```bash
# 觸發更新（改 image）
kubectl set image deployment/my-app api=my-app:v1.3 -n my-app

# 查看更新進度
kubectl rollout status deployment/my-app -n my-app

# 查看更新歷史
kubectl rollout history deployment/my-app -n my-app

# 回滾到上一版
kubectl rollout undo deployment/my-app -n my-app

# 回滾到特定版本
kubectl rollout undo deployment/my-app --to-revision=3 -n my-app

# 暫停更新（出問題時）
kubectl rollout pause deployment/my-app -n my-app

# 繼續更新
kubectl rollout resume deployment/my-app -n my-app
```

---

## 資源 Requests 與 Limits

```yaml
resources:
  requests:    # 排程保證 — Pod 一定有這麼多資源
    cpu: "100m"      # 100 毫核 = 0.1 核
    memory: "128Mi"  # 128 MiB
  limits:      # 上限 — 超過會被限制或殺掉
    cpu: "500m"      # 超過 CPU → 被限速（throttle）
    memory: "256Mi"  # 超過 Memory → 被 OOM Kill
```

| 概念 | 作用 | 沒設的後果 |
|------|------|-----------|
| **requests** | K8s 排程依據，保證最低資源 | Pod 可能被排到資源不足的 Node |
| **limits** | 防止 Pod 吃光整台機器 | 一個 Pod 記憶體洩漏 → 整台 Node 掛掉 |

**最佳實踐**：

- `requests` ≈ 應用正常時的用量
- `limits` ≈ 應用峰值的 1.5-2 倍
- Memory limits 不要設太低（OOM Kill 很痛苦）
- CPU limits 可以設寬鬆一點（限速比殺掉好）

---

## kubectl 常用指令速查

| 指令 | 說明 |
|------|------|
| `kubectl get pods -n my-app` | 查看 Pod 列表 |
| `kubectl get pods -o wide` | 查看 Pod 的 Node 和 IP |
| `kubectl describe pod <name>` | 查看 Pod 詳細資訊（Events 超重要）|
| `kubectl logs <pod> -f` | 查看即時日誌 |
| `kubectl logs <pod> --previous` | 查看上一次 crash 的日誌 |
| `kubectl exec -it <pod> -- sh` | 進入 Pod 的 shell |
| `kubectl port-forward svc/my-app-svc 8080:80` | 本地 port forwarding |
| `kubectl apply -f deployment.yaml` | 套用設定 |
| `kubectl delete -f deployment.yaml` | 刪除資源 |
| `kubectl top pods -n my-app` | 查看 Pod 資源用量 |
| `kubectl get events -n my-app --sort-by=.lastTimestamp` | 查看事件（除錯神器）|
| `kubectl rollout undo deployment/my-app` | 回滾上一版 |

---

## 搶票系統 K8s 部署範例

```
                            Internet
                               │
                        ┌──────▼──────┐
                        │   Ingress   │
                        │  /api → api │
                        │  /ws → ws   │
                        └───┬─────┬───┘
                            │     │
                ┌───────────▼┐   ┌▼───────────┐
                │ api-svc    │   │  ws-svc     │
                │ (ClusterIP)│   │ (ClusterIP) │
                └───┬────────┘   └───┬────────┘
                    │                │
              ┌─────▼─────┐   ┌─────▼─────┐
              │ API Pods  │   │  WS Pods  │
              │ (HPA 2-10)│   │ (HPA 2-5) │
              └─────┬─────┘   └───────────┘
                    │
          ┌────────┼────────┐
          │        │        │
    ┌─────▼──┐ ┌──▼─────┐ ┌▼──────┐
    │PostgreSQL│ │ Redis  │ │ Redis │
    │ (StatefulSet)│ (Queue)│ (Cache)│
    └────────┘ └────────┘ └───────┘
```

```yaml
# 搶票 API Deployment 重點設定
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0       # 搶票期間不能少 Pod
  template:
    spec:
      containers:
      - name: ticket-api
        image: ticket-system:v2.1
        resources:
          requests:
            cpu: "250m"        # 搶票 API 需要較多 CPU
            memory: "256Mi"
          limits:
            cpu: "1000m"
            memory: "512Mi"
        readinessProbe:        # 確保 Redis + DB 都連上才接流量
          httpGet:
            path: /readyz
            port: 8080
```

---

## 本地開發工具

```bash
# 本地 K8s 叢集
minikube start               # Minikube（最簡單）
kind create cluster           # Kind（Docker 內執行，更輕量）
k3d cluster create            # k3d（k3s in Docker）

# 快速部署
helm install my-app ./chart   # Helm Chart（K8s 的套件管理器）
```

---

## FAQ

### Q1：Pod 一直 CrashLoopBackOff 怎麼辦？

先看日誌 `kubectl logs <pod> --previous`，再看事件 `kubectl describe pod <pod>`。常見原因：應用啟動失敗、缺少環境變數、Health Check 設太嚴格。

### Q2：什麼時候用 Deployment，什麼時候用 StatefulSet？

無狀態服務（API Server）用 Deployment，有狀態服務（資料庫、Redis）用 StatefulSet。差別是 StatefulSet 保證 Pod 名稱和儲存空間的穩定性。

### Q3：Ingress 和 LoadBalancer Service 的差別？

LoadBalancer 每個 Service 一個公網 IP（很貴），Ingress 是一個入口依照路徑/域名路由到不同 Service（省錢又靈活）。生產環境幾乎都用 Ingress。

### Q4：replicas 應該設多少？

至少 2（一個掛了還有一個），搭配 HPA 動態調整。搶票系統高峰期可能 10+，平時 2-3 個就夠。

### Q5：為什麼 Pod 被 Evicted（驅逐）？

Node 資源不足時，K8s 會驅逐低優先級 Pod。設定合理的 `requests` 可以減少被驅逐的機率。用 `kubectl describe node` 看 Node 的資源壓力。

---

## 練習

1. 修改 `deployment.yaml`，加入 Readiness Probe（檢查 `/healthz`）和 Liveness Probe
2. 設定 HPA（Horizontal Pod Autoscaler）：CPU 使用率超過 70% 時自動擴展到最多 5 個 Pod
3. 建立一個 ConfigMap 存放 `config.yaml`，掛載到 Pod 中
4. 建立一個 Secret 存放 `JWT_SECRET`，用環境變數注入 Pod
5. 用 `kubectl rollout undo` 練習回滾到上一個版本

---

## 下一課預告

**第三十九課：Circuit Breaker（熔斷器）** — 當下游服務掛掉時，不要傻傻地一直重試，用熔斷器快速失敗、保護系統。
