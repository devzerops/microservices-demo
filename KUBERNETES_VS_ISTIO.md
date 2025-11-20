# Kubernetes Native vs Istio: Resilience Features Comparison

## 핵심 원칙: **함께 사용하는 것이 Best Practice**

Kubernetes Native 기능과 Istio는 서로 **보완 관계**이지 대체 관계가 아닙니다.

| 레벨 | Kubernetes Native | Istio Service Mesh |
|------|-------------------|-------------------|
| **Infrastructure** | ✅ PodDisruptionBudget, HPA | - |
| **Pod Health** | ✅ Readiness/Liveness Probes | - |
| **Request/Network** | - | ✅ Circuit Breaker, Retry |

---

## 1. Circuit Breaker (장애 격리)

### 📊 비교표

| 항목 | Kubernetes Readiness Probe | Istio DestinationRule |
|------|---------------------------|----------------------|
| **동작 레벨** | Pod 전체 | 개별 요청 |
| **격리 단위** | Pod 제거 (0 or 1) | 점진적 차단 (0-100%) |
| **복구** | 수동 또는 HPA | 자동 (시간 기반) |
| **세밀함** | 낮음 | 높음 |
| **설정 필요** | Kubernetes만 | Istio 필요 |

### Kubernetes Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /_readyz
  failureThreshold: 3  # 3회 실패 → Pod 제거
```

**동작 방식:**
1. Health check 3회 연속 실패
2. **Pod 전체를 Service에서 제거**
3. 트래픽 0%로 감소 (All or Nothing)
4. 복구되면 다시 추가

**장점:**
- ✅ Kubernetes 네이티브 (Istio 불필요)
- ✅ 간단하고 확실함

**단점:**
- ❌ 세밀한 제어 불가
- ❌ 일시적 오류에도 Pod 제거
- ❌ 복구 시간이 느림 (initialDelaySeconds)

### Istio Circuit Breaker

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: productcatalogservice
spec:
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 5
      baseEjectionTime: 30s
      maxEjectionPercent: 50%
```

**동작 방식:**
1. 개별 요청 5회 연속 실패
2. **해당 Pod만 30초간 격리** (점진적)
3. 최대 50%까지만 격리 (가용성 보장)
4. 30초 후 자동 복구 시도

**장점:**
- ✅ 세밀한 제어 (요청 레벨)
- ✅ 자동 복구
- ✅ 점진적 격리 (일부 트래픽 유지)

**단점:**
- ❌ Istio 설치 필요
- ❌ 설정이 복잡함

### ✅ Best Practice: 함께 사용

```yaml
# 1단계: Kubernetes Readiness Probe (Pod 레벨)
readinessProbe:
  httpGet:
    path: /_readyz
  failureThreshold: 3

# 2단계: Istio Circuit Breaker (요청 레벨)
outlierDetection:
  consecutiveErrors: 5
  baseEjectionTime: 30s
```

**동작 시나리오:**

```
일시적 오류 (1-4회):
  → Istio가 처리 (30초 격리)
  → Readiness Probe는 통과
  → Pod 유지됨

지속적 오류 (5회+):
  → Istio가 격리
  → Readiness Probe도 실패
  → Pod 제거 또는 재시작
```

---

## 2. Retry (재시도)

### 📊 비교표

| 항목 | Kubernetes Liveness Probe | Istio VirtualService |
|------|--------------------------|---------------------|
| **동작 레벨** | Pod 재시작 | 요청 재시도 |
| **대상** | 전체 프로세스 | 개별 HTTP/gRPC 요청 |
| **백오프** | 없음 | Exponential backoff |
| **오버헤드** | 높음 (Pod 재시작) | 낮음 (요청만) |
| **설정 필요** | Kubernetes만 | Istio 필요 |

### Kubernetes Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /_healthz
  failureThreshold: 3  # 3회 실패 → Pod 재시작
```

**동작 방식:**
1. Health check 3회 연속 실패
2. **Pod 전체를 재시작** (SIGTERM → SIGKILL)
3. 모든 연결 종료
4. 새 Pod 시작 (initialDelaySeconds 대기)

**장점:**
- ✅ Kubernetes 네이티브
- ✅ 완전한 복구 (메모리 누수 등)

**단점:**
- ❌ 오버헤드 큼 (재시작 비용)
- ❌ 일시적 오류에 과함
- ❌ 다운타임 발생

### Istio Retry

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: productcatalogservice
spec:
  http:
  - retries:
      attempts: 3
      perTryTimeout: 2s
      retryOn: "5xx,reset,connect-failure"
```

**동작 방식:**
1. 개별 요청 실패 (5xx, connection failure)
2. **즉시 재시도** (최대 3회)
3. 각 시도마다 2초 타임아웃
4. 성공하면 즉시 반환

**장점:**
- ✅ 빠른 복구 (밀리초 단위)
- ✅ 일시적 오류 자동 처리
- ✅ 사용자에게 투명

**단점:**
- ❌ Istio 설치 필요
- ❌ 영구적 오류는 해결 못함

### ✅ Best Practice: 함께 사용

```yaml
# 1단계: Istio Retry (요청 레벨)
retries:
  attempts: 3
  perTryTimeout: 2s

# 2단계: Liveness Probe (Pod 레벨)
livenessProbe:
  failureThreshold: 3
```

**동작 시나리오:**

```
일시적 네트워크 오류:
  → Istio가 3회 재시도
  → 성공하면 사용자는 모름
  → Liveness Probe는 통과

프로세스 데드락/메모리 누수:
  → Istio 재시도로 해결 안 됨
  → Liveness Probe 3회 실패
  → Pod 재시작으로 근본 해결
```

---

## 3. Zero-Downtime Deployment

### 📊 비교표

| 항목 | Kubernetes PodDisruptionBudget | Istio Circuit Breaker |
|------|-------------------------------|----------------------|
| **보호 대상** | Voluntary disruptions | Application failures |
| **시나리오** | kubectl drain, 업그레이드 | 서비스 장애 |
| **보장** | minAvailable 강제 | Best effort |

### Kubernetes PodDisruptionBudget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: frontend-pdb
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: frontend
```

**동작 방식:**
```
kubectl drain node1
  ↓
PDB 확인: frontend pod가 2개 있나?
  ↓
YES → 1개 종료 허용 (1개는 유지)
NO  → 종료 거부 (minAvailable 보장)
```

**보호하는 시나리오:**
- ✅ `kubectl drain` (노드 정비)
- ✅ 클러스터 업그레이드
- ✅ Voluntary disruption

**보호 못하는 시나리오:**
- ❌ 노드 크래시 (involuntary)
- ❌ 애플리케이션 장애
- ❌ OOM killer

### Istio Circuit Breaker (다시)

애플리케이션 레벨 장애를 보호:
- ✅ 서비스 응답 느림
- ✅ 5xx 에러
- ✅ Connection timeout

### ✅ Best Practice: 함께 사용

```yaml
# Infrastructure 레벨: PDB
minAvailable: 1

# Application 레벨: Istio Circuit Breaker
consecutiveErrors: 5
```

**완전한 보호:**
```
노드 정비 (kubectl drain):
  → PDB가 보호
  → 1개 Pod는 항상 유지

서비스 장애:
  → Istio Circuit Breaker가 보호
  → 장애 Pod 격리

결과: 어떤 상황에서도 가용성 유지
```

---

## 4. Auto-Scaling

### Kubernetes HorizontalPodAutoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: frontend-hpa
spec:
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70
```

**Istio와의 관계:**
- Istio는 스케일링 안 함
- HPA는 **Kubernetes Native 기능**
- Istio Metrics를 HPA에 사용 가능 (고급)

**함께 동작:**
```
트래픽 증가
  ↓
CPU 70% 초과
  ↓
HPA가 Pod 10개로 증가
  ↓
Istio가 10개 Pod에 트래픽 분산
  ↓
부하 분산
```

---

## 전체 아키텍처: Kubernetes + Istio

```
┌─────────────────────────────────────────────────────────┐
│                     사용자 요청                          │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  Istio VirtualService (Retry)                           │
│  - attempts: 3                                          │
│  - perTryTimeout: 2s                                    │
│  → 일시적 오류 자동 복구                                  │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  Istio DestinationRule (Circuit Breaker)                │
│  - consecutiveErrors: 5                                 │
│  - baseEjectionTime: 30s                                │
│  → 장애 Pod 자동 격리                                    │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│  Kubernetes Service (Load Balancing)                    │
│  → Readiness가 통과한 Pod에만 트래픽 전송                │
└─────────────────────────────────────────────────────────┘
                            ↓
        ┌───────────┬───────────┬───────────┐
        ↓           ↓           ↓           ↓
    ┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐
    │Pod 1│    │Pod 2│    │Pod 3│    │Pod 4│
    └─────┘    └─────┘    └─────┘    └─────┘
       ↑           ↑           ↑           ↑
    Readiness  Readiness  Readiness  Readiness
    /_readyz   /_readyz   /_readyz   /_readyz
       ↑           ↑           ↑           ↑
    Liveness   Liveness   Liveness   Liveness
    /_healthz  /_healthz  /_healthz  /_healthz

┌─────────────────────────────────────────────────────────┐
│  PodDisruptionBudget                                    │
│  - minAvailable: 1                                      │
│  → kubectl drain 시에도 1개 Pod 유지                     │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  HorizontalPodAutoscaler                                │
│  - CPU > 70% → Pod 증가                                 │
│  → 부하 증가 시 자동 스케일링                            │
└─────────────────────────────────────────────────────────┘
```

---

## 배포 시나리오별 설정

### 시나리오 1: Istio 없음 (Kubernetes만)

```bash
helm install myboutique ./helm-chart \
  --set resilience.podDisruptionBudget.enabled=true \
  --set resilience.horizontalPodAutoscaler.enabled=true \
  --set resilience.enhancedProbes.enabled=true
```

**보호:**
- ✅ Infrastructure (PDB)
- ✅ Pod health (Probes)
- ✅ Auto-scaling (HPA)
- ❌ Request-level retry/circuit breaking

### 시나리오 2: Istio 사용 (모든 기능)

```bash
helm install myboutique ./helm-chart \
  --set istio.enabled=true \
  --set istio.circuitBreaker.enabled=true \
  --set istio.retry.enabled=true \
  --set resilience.podDisruptionBudget.enabled=true \
  --set resilience.horizontalPodAutoscaler.enabled=true \
  --set resilience.enhancedProbes.enabled=true
```

**보호:**
- ✅ Infrastructure (PDB)
- ✅ Pod health (Probes)
- ✅ Auto-scaling (HPA)
- ✅ Request-level retry (Istio)
- ✅ Circuit breaking (Istio)

---

## 장애 시나리오별 동작

### 1. 일시적 네트워크 오류

```
요청 → ProductCatalog 타임아웃
```

**Istio 있음:**
```
Istio VirtualService
  → 자동으로 3회 재시도
  → 2번째 시도 성공
  → 사용자는 모름 ✅
```

**Istio 없음:**
```
타임아웃 에러
  → 사용자에게 에러 반환 ❌
```

### 2. Pod 장애 (5xx 에러 반복)

```
Pod 1이 계속 5xx 반환
```

**Istio 있음:**
```
5회 연속 에러
  → Istio가 Pod 1 격리 (30초)
  → Pod 2, 3으로 트래픽 전환
  → 사용자는 정상 응답 ✅
```

**Istio 없음:**
```
Readiness Probe 3회 실패
  → Pod 1 제거
  → Pod 2, 3으로 트래픽 전환
  → 더 오래 걸림 (10-30초) ⚠️
```

### 3. 노드 정비 (kubectl drain)

```
kubectl drain node1
```

**PDB 있음:**
```
PDB 확인
  → minAvailable: 1 체크
  → 2개 이상 있으면 1개 종료 허용
  → 무중단 배포 ✅
```

**PDB 없음:**
```
모든 Pod 동시 종료 가능
  → 일시적 다운타임 ❌
```

### 4. 트래픽 급증

```
정상 트래픽의 10배
```

**HPA 있음:**
```
CPU 70% 초과
  → HPA가 Pod 10개로 증가
  → 부하 분산
  → 정상 응답 유지 ✅
```

**HPA 없음:**
```
CPU 100% 도달
  → 응답 속도 저하
  → 일부 요청 타임아웃 ❌
```

---

## 결론

### ✅ Best Practice

```yaml
# 1. Kubernetes Native (필수)
resilience:
  podDisruptionBudget.enabled: true
  horizontalPodAutoscaler.enabled: true
  enhancedProbes.enabled: true

# 2. Istio (권장 - 있으면 더 좋음)
istio:
  enabled: true
  circuitBreaker.enabled: true
  retry.enabled: true
```

### 레이어별 역할

| 레벨 | 기술 | 역할 |
|------|------|------|
| **Infrastructure** | PDB | 노드 정비 시 가용성 보장 |
| **Pod** | Probes | 건강한 Pod만 트래픽 받음 |
| **Pod Scaling** | HPA | 부하에 따라 Pod 수 조절 |
| **Request** | Istio Retry | 일시적 오류 자동 복구 |
| **Request** | Istio Circuit Breaker | 장애 Pod 자동 격리 |

### 왜 함께 사용하나?

1. **Defense in Depth** (다층 방어)
   - 각 레벨에서 다른 종류의 장애 처리
   - 한 레벨 실패해도 다른 레벨이 보호

2. **Fail Fast, Fail Safe**
   - Istio: 빠르게 실패 감지 → 빠른 복구
   - Kubernetes: 확실한 격리 → 안전한 복구

3. **최적화된 복구**
   - 일시적 오류: Istio가 밀리초 단위로 처리
   - 영구적 오류: Kubernetes가 Pod 재시작

**함께 사용하면 99.99% 가용성 달성 가능!** 🎉
