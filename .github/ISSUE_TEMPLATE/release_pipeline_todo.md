---
name: "🚀 Feature & Release TODO Tracker"
about: "Track development tasks, CI commit verification, Go-Doc registration, and PyPI tag release deployment."
title: "[TODO] 모션 파이프라인 개발 및 CI/CD 릴리즈 계획"
labels: ["enhancement", "ci/cd", "todo"]
assignees: ""
---

## 📌 개요 (Overview)
본 이슈는 **Load-Motion-Animation** 엔진 개발, 커밋별 테스트/빌드 자동 검증, Go-Doc 문서 등록 및 PyPI 태그 배포 자동화를 추적하기 위한 TODO 이슈입니다.

---

## 🛠️ TODO 체크리스트 (Tasks & CI/CD Pipeline)

### 1. 🧪 커밋별 테스트 및 빌드 검증 (CI Verification on Every Commit)
- [ ] **Go 단위 테스트 & Vet 검증**: 모든 커밋 마다 `go vet ./...` 및 `go test -v ./tests/...` 통과 필수 ([ci.yml](file:///.github/workflows/ci.yml))
- [ ] **Go 바이너리 빌드 검증**: `go build -o bin/motion-parser ./cmd/motion-parser` 정상 컴파일 확인
- [ ] **Python 단위 테스트 검증**: `PYTHONPATH=python python -m unittest tests/test_python_loader.py` 통과
- [ ] **Python 패키지 빌드 검증**: `python -m build python/` sdist & wheel 정상 생성 확인

### 2. 📚 Go Documentation (Go-Doc) 등록
- [ ] **`pkg.go.dev` 자동 색인 등록**: Go 모듈 주소 (`github.com/Mireu-Lab/Load-motion-animation`) 공개 문서 인덱싱 확인
- [ ] **Go 코드 주석 & 패키지 문서화**: 함수 및 데이터 구조체에 대한 GoDoc 주석 최신화

### 3. 📦 PyPI 배포 (PyPI Package Release)
- [ ] **Python 패키지 설정**: `python/setup.py` 메타데이터, 버전 정보 및 `entry_points` 점검
- [ ] **PyPI 자동 업로드**: 태그 생성 시 `pypa/gh-action-pypi-publish`를 통한 PyPI 자동 배포 및 버전 배포

### 4. 🏷️ 태그 릴리즈 자동화 (Tag Triggered Release)
- [ ] **Git Tag 생성**: `v*` 버전 태그 푸시 (예: `git tag v0.1.0 && git push origin v0.1.0`)
- [ ] **크로스 플랫폼 바이너리 빌드**: Linux(amd64, arm64), macOS(amd64, arm64), Windows(amd64) 멀티 아키텍처 바이너리 자동 생성 ([release.yml](file:///.github/workflows/release.yml))
- [ ] **GitHub Release 생성**: 태그에 컴파일된 바이너리 아티팩트 및 릴리즈 노트 자동 첨부

---

## 💡 연동 워크플로우 (Automated Workflows)
- [CI Workflow](file:///.github/workflows/ci.yml): 매 Push / Pull Request 커밋 단위 빌드 & 테스트 자동화
- [Release Workflow](file:///.github/workflows/release.yml): `v*` Tag 감지 시 PyPI 배포, Go-Doc 인덱싱, GitHub Release 게시
- [Auto Branch Workflow](file:///.github/workflows/create-issue-branch.yml): Issue 생성 시 작업 브랜치 자동생성 및 Push
