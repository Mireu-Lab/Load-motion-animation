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
    "load-motion-animation/pkg/bvh"
    "load-motion-animation/pkg/df"
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
# Go 테스트
go test -v ./tests

# Python 테스트
PYTHONPATH=python python3 -m unittest tests/test_python_loader.py
```
