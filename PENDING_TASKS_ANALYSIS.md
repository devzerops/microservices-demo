# 추가 작업 분석 보고서 (Pending Tasks Analysis Report)

**프로젝트**: Online Boutique - Google Cloud Platform 마이크로서비스 데모
**분석 날짜**: 2025-11-08
**분석 범위**: 전체 코드베이스

---

## 📋 목차

1. [주요 미완성 작업 (High Priority)](#주요-미완성-작업-high-priority)
2. [코드 품질 개선 (Code Quality Improvements)](#코드-품질-개선-code-quality-improvements)
3. [보안 및 최적화 (Security & Optimization)](#보안-및-최적화-security--optimization)
4. [문서화 및 배포 (Documentation & Deployment)](#문서화-및-배포-documentation--deployment)
5. [의존성 업데이트 (Dependency Updates)](#의존성-업데이트-dependency-updates)
6. [테스트 개선 (Test Improvements)](#테스트-개선-test-improvements)

---

## 주요 미완성 작업 (High Priority)

### 1. OpenTelemetry 구현 미완성 ⚠️

여러 Go 마이크로서비스에서 OpenTelemetry 통계 및 추적 기능이 미구현 상태입니다.

#### 영향받는 서비스:

**Frontend Service** (`src/frontend/main.go:173`)
```go
func initStats(log logrus.FieldLogger) {
    // TODO(arbrown) Implement OpenTelemtry stats
}
```

**Shipping Service** (`src/shippingservice/main.go:150,154`)
```go
func initStats() {
    //TODO(arbrown) Implement OpenTelemetry stats
}

func initTracing() {
    // TODO(arbrown) Implement OpenTelemetry tracing
}
```

**Checkout Service** (`src/checkoutservice/main.go:149`)
```go
func initStats(log logrus.FieldLogger) {
    //TODO(arbrown) Implement OpenTelemetry stats
}
```

**Product Catalog Service** (`src/productcatalogservice/server.go:151`)
```go
// TODO(drewbr) Implement OpenTelemetry stats
```

**작업 내용**:
- 각 서비스에 OpenTelemetry 통계 수집 구현
- Tracing 기능 완성
- 메트릭 수집 및 전송 로직 추가

---

### 2. Stackdriver Profiler 비활성화 (#2517) ⚠️

Ad Service에서 Stackdriver Profiler가 주석 처리되어 있습니다.

**위치**:
- `src/adservice/Dockerfile:30`
- `src/adservice/build.gradle:100,110`

**관련 이슈**: https://github.com/GoogleCloudPlatform/microservices-demo/issues/2517

**작업 내용**:
```dockerfile
# @TODO: https://github.com/GoogleCloudPlatform/microservices-demo/issues/2517
# Download Stackdriver Profiler Java agent
# RUN mkdir -p /opt/cprof && \
```

```gradle
// @TODO: https://github.com/GoogleCloudPlatform/microservices-demo/issues/2517
// defaultJvmOpts =
//          ["-agentpath:/opt/cprof/profiler_java_agent.so=-cprof_service=adservice,-cprof_service_version=1.0.0"]
```

**필요 작업**:
- Issue #2517 검토 및 해결
- Stackdriver Profiler 재활성화 또는 대안 마련

---

### 3. Shopping Assistant Service - Helm 지원 미구현 ⚠️

**위치**: `helm-chart/values.yaml:216`

```yaml
# @TODO: This service is not currently available in Helm.
# https://github.com/GoogleCloudPlatform/microservices-demo/tree/main/kustomize/components/shopping-assistant
shoppingAssistantService:
```

**작업 내용**:
- Shopping Assistant Service의 Helm Chart 구현
- Kustomize 구성을 Helm으로 변환
- 관련 테스트 추가

---

## 코드 품질 개선 (Code Quality Improvements)

### 1. 중복 코드 제거 - 로깅 클래스 🔄

**Python 서비스 간 로거 클래스 중복**

**위치**:
- `src/emailservice/logger.py:21`
- `src/recommendationservice/logger.py:21`

```python
# TODO(yoshifumi) this class is duplicated since other Python services are
# not sharing the modules for logging.
class CustomJsonFormatter(jsonlogger.JsonFormatter):
    # ...
```

**작업 내용**:
- 공통 로깅 모듈 생성 (예: `src/common/python/logger.py`)
- 모든 Python 서비스에서 공통 모듈 사용하도록 리팩토링
- 패키징 및 의존성 관리 설정

---

### 2. 중복 코드 제거 - Profiling 초기화 🔄

**Go 서비스 간 Profiling 초기화 함수 중복**

**위치**:
- `src/shippingservice/main.go:158`
- `src/frontend/main.go:194`
- `src/checkoutservice/main.go:179`

```go
// TODO(ahmetb) this method is duplicated in other microservices using Go
// since they are not sharing packages.
func initProfiling(service, version string) {
    // ...
}
```

**작업 내용**:
- 공통 Go 패키지 생성 (예: `pkg/common/profiling`)
- 모든 Go 서비스에서 공통 패키지 사용
- Go modules 설정 업데이트

---

## 보안 및 최적화 (Security & Optimization)

### 1. AlloyDB 보안 개선 🔒

**위치**: `src/cartservice/src/cartstore/AlloyDBCartStore.cs:42,46`

```csharp
// TODO: Create a separate user for connecting within the application
// rather than using our superuser
string alloyDBUser = "postgres";

// TODO: Consider splitting workloads into read vs. write and take
// advantage of the AlloyDB read pools
```

**작업 내용**:
- AlloyDB용 별도 애플리케이션 사용자 생성
- Superuser 대신 제한된 권한의 사용자 사용
- Read/Write 워크로드 분리
- AlloyDB Read Pool 활용

---

### 2. 프론트엔드 UI 복원 🎨

**위치**: `src/frontend/templates/home.html:35`

```html
<!-- @TODO: removed temporarily. When uncommenting, also replace below div with this -->
<!--<div class="col-4 d-none d-lg-block home-desktop-left-image"></div>-->
```

**작업 내용**:
- 홈페이지 왼쪽 이미지 복원 검토
- 레이아웃 조정 완료
- 반응형 디자인 테스트

---

## 문서화 및 배포 (Documentation & Deployment)

### 1. DeployStack 브랜치 참조 업데이트 📚

**위치**: `docs/deploystack.md:5`

```markdown
<!-- TODO: remove reference to the deploystack-enable branch when it pushes to main -->
```

**작업 내용**:
- deploystack-enable 브랜치가 main에 병합되었는지 확인
- 문서에서 브랜치 참조 제거
- DeployStack 문서 업데이트

---

### 2. ProductCatalogService 버그 문서화 🐛

**위치**: `src/productcatalogservice/README.md:14,23,27`

```markdown
However, this feature is bugged: the catalog is actually reloaded on each

# Trigger bug
...
# Remove bug
```

**작업 내용**:
- 카탈로그 리로드 버그 수정
- 파일 감시 기능 개선
- 관련 테스트 추가

---

## 의존성 업데이트 (Dependency Updates)

### 1. Deprecated OpenTelemetry 패키지 ⚠️

**Payment Service** (`src/paymentservice/package-lock.json`)

다음 패키지들이 deprecated 상태입니다:

```json
"deprecated": "Please use @opentelemetry/api >= 1.3.0"
"deprecated": "Please use trace and metric specific exporters @opentelemetry/exporter-trace-otlp-grpc and @opentelemetry/exporter-metrics-otlp-grpc"
"deprecated": "Please use trace and metric specific exporters @opentelemetry/exporter-trace-otlp-http and @opentelemetry/exporter-metrics-otlp-http"
"deprecated": "Please use @opentelemetry/sdk-metrics"
```

**작업 내용**:
- OpenTelemetry API를 1.3.0 이상으로 업데이트
- 트레이스 및 메트릭 전용 exporter로 변경
- sdk-metrics 패키지로 마이그레이션
- 호환성 테스트 수행

---

### 2. Renovate 자동 업데이트 설정 검토 🔄

**위치**: `.github/renovate.json5`

현재 설정:
- Python 버전: `~=3.11.0`
- Kubernetes manifest 제외 경로: `release/**`, `kustomize/base/**`
- 스케줄: 월요일 이른 시간

**작업 내용**:
- Python 3.12+ 마이그레이션 고려
- Renovate PR 검토 및 병합
- 자동 업데이트 정책 재검토

---

## 테스트 개선 (Test Improvements)

### 1. Shipping Service 테스트 품질 향상 🧪

**위치**: `src/shippingservice/shippingservice_test.go:86`

```go
// @todo improve quality of this test to check for a pattern such as '[A-Z]{2}-\d+-\d+'.
if len(res.TrackingId) != 18 {
    t.Errorf("TestShipOrder: Tracking ID is malformed - has %d characters, %d expected", len(res.TrackingId), 18)
}
```

**작업 내용**:
- 정규표현식을 사용한 Tracking ID 패턴 검증 추가
- 형식: `[A-Z]{2}-\d+-\d+` (예: AB-12345-67890)
- Edge case 테스트 추가
- 테스트 커버리지 향상

**개선 예시**:
```go
trackingIDPattern := regexp.MustCompile(`^[A-Z]{2}-\d+-\d+$`)
if !trackingIDPattern.MatchString(res.TrackingId) {
    t.Errorf("TestShipOrder: Tracking ID doesn't match expected pattern [A-Z]{2}-\\d+-\\d+, got: %s", res.TrackingId)
}
```

---

### 2. 테스트 커버리지 확대 필요 📊

현재 확인된 테스트 파일:
- `src/shippingservice/shippingservice_test.go`
- `src/productcatalogservice/product_catalog_test.go`
- `src/checkoutservice/money/money_test.go`
- `src/frontend/money/money_test.go`
- `src/frontend/validator/validator_test.go`

**작업 내용**:
- 각 마이크로서비스별 통합 테스트 추가
- E2E 테스트 시나리오 확대
- 테스트 커버리지 80% 이상 목표
- CI/CD 파이프라인에 커버리지 리포트 추가

---

## 🎯 우선순위 권장사항

### High Priority (즉시 착수)
1. ⚠️ OpenTelemetry 구현 완료 (모든 Go 서비스)
2. ⚠️ AlloyDB 보안 개선 (superuser 사용 중지)
3. ⚠️ Deprecated OpenTelemetry 패키지 업데이트

### Medium Priority (단기 계획)
4. 🔄 중복 코드 제거 (Python logger, Go profiling)
5. 🔒 Stackdriver Profiler 이슈 #2517 해결
6. 🎨 Shopping Assistant Helm 지원 추가

### Low Priority (장기 계획)
7. 📚 문서화 개선 (DeployStack, ProductCatalog 버그)
8. 🧪 테스트 품질 및 커버리지 향상
9. 🎨 프론트엔드 UI 복원

---

## 📝 추가 권장사항

### CI/CD 개선
- GitHub Actions 워크플로우에 테스트 커버리지 리포트 추가
- 보안 스캔 자동화 (Dependabot, Snyk 등)
- 성능 테스트 자동화

### 모니터링 및 관찰성
- OpenTelemetry 완전 구현 후 Grafana 대시보드 구성
- 분산 추적 설정 및 검증
- 로그 집계 시스템 개선

### 보안
- 모든 서비스의 의존성 취약점 스캔
- 비밀 관리 개선 (Secret Manager 활용)
- RBAC 정책 강화

---

## 📊 통계 요약

- **총 TODO 항목**: 15개
- **고우선순위**: 5개
- **중우선순위**: 4개
- **저우선순위**: 6개
- **영향받는 서비스**: 11개 중 8개
- **주요 언어**: Go (5), Python (2), C# (1), Java (1)

---

**보고서 끝**
