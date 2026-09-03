# Load-motion-animation 🎬

> **다양한 모션 파서 및 데이터 파이프라인 엔진 (Python & Golang)**  
> 3D/2D 모션 Capture 및 3D 애니메이션 포맷 데이터를 손쉽게 로드하고, 데이터베이스 및 분석용 데이터 프레임워크(DataFrame) 형태로 변환·저장하기 위한 프로젝트입니다.

---

## 📌 1. 프로젝트 배경 및 목적 (Background & Objectives)

3D 모션 생성 AI 모델 연구 및 모션 데이터 처리 작업 시, 기존 Python이나 Golang 생태계에서 다양한 3D/2D 모션 데이터 포맷을 통합적으로 파싱하고 처리하는 호환성 높은 라이브러리가 부족하다는 문제점을 해결하기 위해 시작되었습니다.

본 프로젝트의 목표는 다음과 같습니다:
- **통합 로더 제공**: 호환성이 파편화되어 있는 다양한 모션 캡처 및 3D 데이터 포맷(`FBX`, `USD`, `BVH`, `C3D`, `TRC` 등)을 단일 파이프라인에서 읽을 수 있도록 지원합니다.
- **다양한 Storage 변환 지원**: 파싱된 모션 데이터를 RDBMS(PostgreSQL, MySQL, SQLite), NoSQL(MongoDB/JSON), CSV/TSV 파일 등 목적에 맞게 손쉽게 내보내고 저장합니다.
- **데이터 분석 및 ML 연동**: 판다스(Pandas) / Polars 등 데이터프레임 구조로 즉시 변환하여 모션 캡처 데이터 분석 및 AI 훈련 데이터셋 구축 속도를 극대화합니다.
- **모션 시각화**: 3D 관절 궤적 및 뼈대 동작 상태를 2D 캔버스에 투영하여 GIF 애니메이션으로 렌더링 내보내기를 지원합니다.

---

## 🛠️ 2. 기술 스택 (Tech Stack & Target Languages)

우리는 실행 언어로 **Golang** 및 **Python** 모듈을 함께 제공합니다.

- 🐍 **Python**: PyTorch / TensorFlow AI 모델 학습 파이프라인, Pandas / Polars 기반 데이터 프레임 지원
- 🐹 **Go (Golang)**: 대용량 모션 파일의 초고속 동시성(Concurrent) 파싱 엔진, CLI 툴킷 및 DB Batch Exporter 제공

---

## 🚀 3. 주요 기능 및 구현 현황 (Features & TODO List)

### 📂 1) Motion File Load Support (Python & Go)
다양한 3D 모션 데이터 포맷 지원 파서 개발:
- [x] **Motion Capture Data**: `BVH`, `C3D`, `TRC` (AMC/ASF, MVN 확장 인터페이스 제공)
- [x] **Calibration & Meta Data**: `CALIB` 메타데이터 구조체 지원

### 💾 2) Data Export & Storage Integration
- [x] **RDBMS**: PostgreSQL, MySQL, SQLite 배치 INSERT SQL 스크립트 생성 지원
- [x] **NoSQL**: MongoDB import용 JSON/NDJSON 문서 Export 지원
- [x] **File System Export**: CSV / TSV 포맷 구조화 데이터 출력 지원

### 📊 3) DataFrame & Analysis Ready
- [x] Python DataFrame (Pandas, Polars, Dict) 및 Go 메모리 DataFrame 직렬화/역직렬화 인터페이스 제공

### 🎨 4) Motion Visualization Export
- [x] **Animated GIF Export**: 3D 관절 궤적 및 뼈대 동작 상태를 2D 캔버스에 투영하여 GIF 애니메이션으로 내보내기 기능 지원

---

## 💻 4. 사용 방법 (Usage Guide)

### 🐹 1) Golang CLI & Package

#### CLI 실행
```bash
# Go 바이너리 빌드
go build -o bin/motion-parser ./cmd/motion-parser

# BVH 모션 파일 분석 및 데이터프레임 요약 출력
./bin/motion-parser -input _sample/Sample_motion.bvh -export summary

# CSV / TSV 내보내기
./bin/motion-parser -input _sample/Sample_motion.bvh -export csv -output output.csv

# PostgreSQL 배치 Insert SQL 파일 생성
./bin/motion-parser -input _sample/Sample_motion.bvh -export sql -db postgresql -output output.sql

# 3D 스켈레톤 GIF 애니메이션 내보내기
./bin/motion-parser -input _sample/Sample_motion.bvh -export gif -output motion.gif -max-frames 120
```


#### Go 코드 예시
```go
package main

import (
    "fmt"
    "gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/bvh"
    "gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/pkg/df"
)

func main() {
    parser := bvh.NewParser()
    motionData, err := parser.ParseFile("_sample/Sample_motion.bvh")
    if err != nil {
        panic(err)
    }

    dataframe := df.FromMotionData(motionData)
    fmt.Println(dataframe.Summary())
}
```

---

### 🐍 2) Python Package (`load_motion`)

#### CLI 실행
```bash
# Python CLI 실행
PYTHONPATH=python python3 -m load_motion.cli -input _sample/Sample_motion.bvh -export summary
PYTHONPATH=python python3 -m load_motion.cli -input _sample/arthuman-sample.c3d -export csv -output sample.csv
```

#### Python 코드 예시
```python
from load_motion import load_motion, to_dataframe, export_csv, export_sql

# 1. 모션 파일 로드 (BVH, C3D, TRC 자동 감지)
motion = load_motion("_sample/Sample_motion.bvh")
print(f"Format: {motion.format}, Frames: {motion.total_frames}, FPS: {motion.frame_rate}")

# 2. Pandas DataFrame 변환
df = to_dataframe(motion, library="pandas")
print(df.head())

# 3. CSV 및 PostgreSQL SQL 생성
export_csv(motion, "motion.csv")
export_sql(motion, "motion.sql", db_type="postgresql")
```

---

## 🧪 5. 테스트 실행 (Run Tests)

```bash
# Go 테스트 및 커버리지
go test -v -cover ./...

# Python 테스트
PYTHONPATH=python python3 -m unittest tests/test_python_loader.py
```

---

## 📖 6. GoDoc 문서화 (Go Documentation)

본 프로젝트의 모든 Go 패키지 및 모듈은 표준 GoDoc 규격에 맞춰 상세한 주석과 Example 테스트 코드를 포함하고 있습니다.

### 1) 터미널에서 빠른 API 문서 조회 (`go doc`)
```bash
# 패키지 개요 조회
go doc ./pkg/motion
go doc ./pkg/bvh

# 특정 구조체 및 메서드 상세 조회
go doc ./pkg/motion.MotionData
go doc ./pkg/bvh.Parser
go doc ./pkg/exporter.ExportSQLScript
go doc ./pkg/visualizer.GIFOptions
```

### 2) 로컬 대화형 GoDoc 웹 서버 실행
스크립트를 실행하면 로컬 웹 서버가 구동되어 브라우저에서 편리하게 API 문서를 열람할 수 있습니다:
```bash
./scripts/run_godoc.sh
# 또는 포트 지정 (예: 6060)
./scripts/run_godoc.sh 6060
```
- 브라우저 접속: `http://localhost:6060/pkg/gitlab.com/none-labs/behavioral-psychology-projects/Load-motion-animation/`

### 3) 정적 HTML GoDoc 문서 생성 (GitLab Pages 배포용)
```bash
# public/ 디렉토리에 정적 사이트 빌드
./scripts/generate_godoc.sh public
```

---

## 🦊 7. GitLab Runner CI/CD 파이프라인 (.gitlab-ci.yml)

GitLab Runner 환경에서 자동으로 빌드, 테스트, 문서화, 배포가 수행되도록 다단계 CI/CD 파이프라인을 지원합니다.

### 🔄 파이프라인 단계 (Stages)

| 단계 (Stage) | Job 이름 | 실행 내용 | 실행 조건 |
| :--- | :--- | :--- | :--- |
| **test** | `test:go` | Go 1.22 환경에서 `go vet`, 단위 테스트, Example 테스트, CLI 빌드 검증 및 테스트 커버리지 리포트 생성 | Push, MR |
| **test** | `test:python` | Python 3.11 환경에서 Go 엔진 아티팩트와 연동한 모션 로더 테스트 및 sdist/wheel 패키지 빌드 | Push, MR |
| **build** | `build:cross-platform` | Linux/macOS/Windows (amd64, arm64) 5종 크로스 컴파일 바이너리 빌드 및 아티팩트 보관 | `main`, Tag |
| **docs** | `verify:godoc` | Merge Request 시 정적 GoDoc 빌드 무결성 사전 검증 | MR |
| **docs** | `pages` | GoDoc 정적 웹 문서를 생성하여 **GitLab Pages** 사이트로 자동 호스팅 배포 | `main`, Tag |
| **release** | `release:gitlab` | 태그 푸시 시 `release-cli`를 통해 GitLab Release 자동 발행 및 크로스 바이너리 첨부 | Tag (`v*`) |
| **release** | `publish:pypi` | 태그 푸시 시 PyPI로 Python 라이브러리 자동 배포 (선택적) | Tag (`v*`) |
| **release** | `trigger:godoc-index` | 태그 푸시 시 `proxy.golang.org` 인덱싱 요청으로 `pkg.go.dev` 공식 문서 갱신 | Tag (`v*`) |

### ⚡ GitLab Runner 설정 최적화
- **Docker Executor**: `golang:1.22`, `python:3.11`, `curlimages/curl` 등 표준 컨테이너 이미지 기반 구동
- **아티팩트 파이프라인 연계**: `test:python` 단계에서 Go 빌드 컴파일러 설치 없이 `test:go`의 `bin/motion-parser` 바이너리 아티팩트를 직접 전달받아 초고속 테스트 실행
- **캐싱 (Caching)**: `.go/pkg/mod/` 및 `.cache/pip/` 경로를 캐싱하여 러너 실행 시간을 대폭 단축
