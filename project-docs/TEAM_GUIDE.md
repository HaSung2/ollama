# 공소 3조 작업 지시서 (팀원 + AI 에이전트용)

> 대상: 3조 팀원 4명과 각 팀원이 사용하는 AI 코딩 에이전트
> 저장소: `https://github.com/HaSung2/ollama` (기본 브랜치 `main`, 기준 커밋 `d63b702`)
> 작성: 팀장 송하성 / 최종 수정 2026-11-22
> 최종 발표: 2026-12-01(화) / 코드 프리즈: 2026-11-28(토) 23:59

---

## 0. 이 문서 사용법

**팀원(사람)**
1. 1~4장을 읽고 환경을 준비한다. 깃허브가 처음이어도 3장만 읽으면 흐름을 이해할 수 있다.
2. 10장 "에이전트 시작 프롬프트"를 복사해서 이름과 담당 섹션만 바꿔 AI 에이전트에게 붙여 넣는다.
3. 에이전트가 PR을 만들고 멈추면, PR 링크를 단톡방에 공유한다. 리뷰와 머지는 팀장이 한다.

**AI 에이전트**
- 이 문서 전체를 읽은 뒤 작업한다. 이 문서의 규칙이 에이전트의 기본 습관(자동 서명, 이모지, squash 등)보다 우선한다.
- 담당 섹션(8장)에 있는 이슈만 처리한다. 7장의 계약(함수 이름, 시그니처, 파일 소유권, 출력 형식, JSON 스키마)은 임의로 바꾸지 않는다.
- 각 단계를 마칠 때마다 팀원에게 "무엇을, 왜 했는지"를 2~3줄로 설명한다. 팀원은 깃허브를 처음 써 보며 배우는 중이다.
- PR을 생성하면 **멈추고 보고한다.** 직접 머지하지 않는다.

---

## 1. 프로젝트 요약과 원칙

Ollama CLI에 두 가지 기능을 추가한다.

| 기능 | 설명 |
|---|---|
| `ollama run <모델> --profile` | 응답을 생성하는 동안 0.2초 간격으로 Ollama 프로세스의 CPU·RAM 사용량과 GPU 사용량을 샘플링한다. 응답이 끝나면 평균값과 최댓값을 `--verbose` 출력 아래에 표시한다. |
| `ollama stats [MODEL]` | `--profile` 측정 결과를 `~/.ollama/stats.jsonl`에 누적 저장하고, 기록 목록과 모델별 평균을 표로 보여준다. |

**지켜야 할 원칙**
1. 수정은 `cmd/` 안에서만 한다. `server/`, `llm/`, `ml/`, `api/types.go`는 수정하지 않는다. 서버는 공식 설치본을 그대로 쓰고, 수정한 CLI만 빌드한다.
2. `go.mod`와 `go.sum`은 팀장만 수정한다(H-1에서 gopsutil 추가).
3. 7장의 계약을 지켜야 4명이 동시에 작업해도 코드가 맞물린다.

---

## 2. 팀장 사전 작업

팀장이 저장소 전환, 설정, 라벨 생성, S-1과 H-1 머지를 별도 문서(`LEADER_SETUP.md`)로 진행한다. 팀원과 에이전트는 이 장을 건너뛰고, 팀장이 완료를 공지하면 4장부터 시작한다.

팀원이 알아 둘 결과만 정리하면 다음과 같다.
- `HaSung2/ollama`는 fork가 아닌 독립 저장소이고, 기본 브랜치는 `main`(기준 커밋 `d63b702`)이다.
- `main`에는 직접 푸시할 수 없고 PR로만 들어간다. 머지 방식은 Create a merge commit이다.
- 라벨(`type:*`, `area:*`, `status:blocked`)과 마일스톤(`최종 발표 (12/1)`)이 미리 만들어져 있다.

---

## 3. 깃·깃허브 기초 (처음 쓰는 팀원용)

### 3.1 용어
| 용어 | 뜻 |
|---|---|
| 저장소(repository) | 코드와 변경 이력이 저장되는 곳. 우리는 `HaSung2/ollama` 하나를 함께 쓴다. |
| 브랜치(branch) | 작업 공간을 갈라놓은 것. `main`은 완성본이고, 각자 새 브랜치에서 작업한다. |
| 커밋(commit) | 변경 사항을 의미 있는 단위로 저장한 기록 1개 |
| 푸시(push) | 내 컴퓨터의 커밋을 GitHub에 올리는 것 |
| 이슈(issue) | 해야 할 일 1개를 적은 카드. 번호(#12)가 붙는다. |
| PR(Pull Request) | "내 브랜치를 main에 합쳐 주세요"라는 요청. 리뷰를 받는 곳이다. |
| 리뷰 / 머지 | 팀장이 PR을 검토하고(리뷰) main에 합친다(머지). |

### 3.2 작업 한 사이클
```
이슈 생성 (#12)
   |
main 최신화 -> 브랜치 생성 (feat/12-profile-flag)
   |
코드 수정 -> 커밋 -> 코드 수정 -> 커밋 ... (작은 단위로 여러 번)
   |
검사 (gofmt, go vet, go test)
   |
push -> PR 생성 (Closes #12, 리뷰어: HaSung2)
   |
팀장 리뷰 -> (수정 요청 시 추가 커밋 후 push) -> 팀장 머지 -> 이슈 자동 종료
   |
main 최신화 -> 다음 이슈
```

---

## 4. 환경 준비 (팀원 공통, 에이전트가 점검)

| 순서 | 항목 | 명령 / 방법 | 확인 기준 |
|---|---|---|---|
| 1 | 초대 수락 | 메일 또는 `https://github.com/HaSung2/ollama/invitations` | 저장소에 push 권한 있음 |
| 2 | Git | macOS: `xcode-select --install`, Windows: Git for Windows | `git --version` |
| 3 | GitHub CLI | macOS: `brew install gh`, Windows: `winget install GitHub.cli` | `gh --version` |
| 4 | Go | https://go.dev/dl (`go.mod`의 `go` 줄 버전 이상) | `go version` |
| 5 | C 컴파일러 | macOS: Xcode CLT(Clang). Windows: TDM-GCC 설치 후 PATH 등록 | `gcc --version` (Windows) |
| 6 | 공식 Ollama | https://ollama.com/download 설치 후 실행 | `ollama --version`, `ollama list` |
| 7 | Git 사용자 설정 | 아래 코드 블록 | 이메일이 GitHub 계정 이메일과 같음 |
| 8 | gh 로그인 | `gh auth login` (GitHub.com, HTTPS, 브라우저 로그인) | `gh auth status` |
| 9 | 저장소 받기 | 아래 코드 블록 | `git remote -v`가 HaSung2/ollama |
| 10 | 빌드 | macOS: `go build -o ollama-dev .` / Windows: `go build -o ollama-dev.exe .` | `./ollama-dev --version` 실행 |

```bash
# 7. 커밋 작성자 설정: 이메일이 GitHub 계정에 등록된 이메일과 다르면 기여도에 집계되지 않는다
git config --global user.name  "본인 이름 또는 GitHub 아이디"
git config --global user.email "GitHub 계정에 등록된 이메일"
# GitHub의 noreply 이메일(…@users.noreply.github.com)을 써도 된다

# 9. 저장소 받기
#    경로에 공백, 대괄호, 한글이 있으면 CGO 빌드가 실패한다. 예: ~/dev/ollama, C:\dev\ollama
git clone https://github.com/HaSung2/ollama.git
cd ollama
gh repo set-default HaSung2/ollama
```

> Ollama에는 CGO(C 코드)가 포함되어 있어 Windows에서도 C 컴파일러가 있어야 `go build`와 `go test`가 된다. CPU 전용 빌드여도 우리 기능은 CLI 쪽이므로 문제없다.
> 빌드 환경이 끝내 안 되면 코드 작성과 PR은 그대로 진행한다. 검증은 CI(S-2)와 팀장 맥북에서 한다. 이 경우 PR의 "테스트" 칸에 "로컬 빌드 불가, CI 결과로 확인"이라고 적는다.

---

## 5. 작업 규칙

### 5.1 브랜치
- 형식: `<type>/<이슈번호>-<짧은-영문-설명>`
- 예: `feat/12-profile-flag`, `test/25-stats-edge-cases`, `docs/30-profile-guide`
- 항상 최신 `main`에서 만든다.
  ```bash
  git switch main && git pull origin main
  git switch -c feat/12-profile-flag
  ```
- `main`에 직접 커밋하거나 푸시하지 않는다. 다른 사람의 브랜치에도 푸시하지 않는다.

### 5.2 커밋 컨벤션 (Conventional Commits)
```
<type>(<scope>): <요약>

<본문: 무엇을, 왜 바꿨는지. 선택이지만 권장>

Refs #<이슈번호>
```

| type | 용도 |
|---|---|
| `feat` | 기능 추가 |
| `fix` | 버그 수정, 리뷰 지적 수정 |
| `test` | 테스트 추가, 수정 |
| `docs` | 문서 |
| `refactor` | 동작 변화 없는 구조 개선 |
| `build` | 의존성, 빌드 설정 |
| `ci` | CI 워크플로 |
| `chore` | 템플릿, 설정 등 기타 |

| scope | 범위 |
|---|---|
| `profile` | 샘플러, 자료구조, 요약 (`cmd/profile*.go`) |
| `gpu` | OS별 GPU 측정 (`cmd/gpu_*.go`) |
| `run` | run 명령 플래그, chat/generate 연결 |
| `interactive` | 대화형 슬래시 명령 |
| `stats` | 기록 저장, stats 명령 |
| `bench` | 측정, 그래프 스크립트 |
| `docs` / `github` / `ci` / `deps` | 문서 / 템플릿 / CI / 의존성 |

**규칙**
1. 요약은 한국어, 50자 이내로 쓰고 마침표를 찍지 않는다. 예: `feat(run): --profile 플래그 등록`
2. **커밋 하나 = 논리적 변경 하나.** 각 이슈의 "커밋 계획"을 따라 PR 하나에 3~7개의 커밋을 만든다. 커밋 시점마다 빌드가 되는 상태를 유지한다.
3. 커밋 수를 늘리려고 의미 없이 쪼개지 않는다. 예: 공백 하나, 오타 하나만 바꾸는 커밋, 같은 파일을 되돌리는 커밋은 금지한다. 리뷰어와 교수님이 이력을 본다.
4. 마지막 줄에 `Refs #이슈번호`를 넣는다.
5. **이모지와 이모티콘 문자를 쓰지 않는다.** AI 도구가 자동으로 붙이는 서명이나 푸터에 이모지가 있으면 지운다. 이모지가 없는 텍스트 `Co-Authored-By:` 줄은 허용한다.
6. 커밋 전 `git diff --staged`로 의도한 파일만 들어갔는지 확인한다. 빌드 결과물(`ollama-dev`, `ollama-dev.exe`), 개인 설정 파일, 측정 원본 로그는 커밋하지 않는다.

### 5.3 이슈
- 담당자는 자기 이슈를 **본인 계정으로** 직접 만든다(본인 기여로 집계된다).
- 제목: `[작업ID] <type>(<scope>): <요약>` 예: `[W-1] feat(run): --profile 플래그 추가`
- 본문: 6.1 템플릿의 모든 칸을 채운다. 8장의 명세를 옮겨 적으면 된다.
- 라벨: `type:*` 1개 + `area:*` 1개, Assignee: 본인, Milestone: `최종 발표 (12/1)`
- 선행 작업이 아직 머지되지 않았으면 `status:blocked` 라벨을 붙이고, 본문 "선행 작업"에 선행 이슈 번호를 적는다.
```bash
gh issue create --repo HaSung2/ollama \
  --title "[W-1] feat(run): --profile 플래그 추가" \
  --body-file issue-W-1.md \
  --label "type:feat" --label "area:profile" \
  --assignee "@me" --milestone "최종 발표 (12/1)"
```
- 본문 파일(`issue-*.md`)은 임시 파일이므로 저장소에 커밋하지 않는다.

### 5.4 PR
- **이슈 1개 = PR 1개.** 여러 이슈를 한 PR에 섞지 않는다.
- base는 `main`, 저장소는 반드시 `HaSung2/ollama`. 원본 `ollama/ollama`에는 절대 이슈나 PR을 만들지 않는다.
- 제목은 이슈 제목과 같다. 본문은 6.3 PR 템플릿을 모두 채우고 `Closes #이슈번호`를 넣는다.
- 리뷰어: `HaSung2`, Assignee: 본인, 라벨과 마일스톤은 이슈와 같게 한다.
- 선행 PR이 아직 머지되지 않아 단독으로 빌드되지 않으면 Draft PR로 올리고 본문에 의존 관계를 적는다.
```bash
git push -u origin feat/12-profile-flag
gh pr create --repo HaSung2/ollama --base main --head feat/12-profile-flag \
  --title "[W-1] feat(run): --profile 플래그 추가" \
  --body-file pr-W-1.md \
  --reviewer HaSung2 --assignee "@me" \
  --label "type:feat" --label "area:profile" --milestone "최종 발표 (12/1)"
```

### 5.5 리뷰와 머지
- 팀장이 리뷰한 뒤 **Create a merge commit** 방식으로 머지한다. 작성자는 직접 머지하지 않는다.
- 수정 요청을 받으면 같은 브랜치에 **새 커밋을 추가**해서 푸시한다(`fix(<scope>): 리뷰 반영 - <내용>`). 리뷰가 시작된 뒤에는 force push, amend, rebase를 하지 않는다.
- 반영 후 리뷰 댓글에 무엇을 고쳤는지 답글을 단다.
- 머지 후 정리:
  ```bash
  git switch main && git pull origin main
  git branch -d feat/12-profile-flag
  ```

### 5.6 충돌이 났을 때
rebase 대신 merge로 해결한다(초보자에게 안전하고, force push가 필요 없다).
```bash
git fetch origin
git merge origin/main        # 충돌 파일 수정
git add <파일> && git commit # 기본 병합 메시지를 그대로 사용
git push
```
충돌 부분이 다른 사람의 담당 코드라면 임의로 결정하지 말고, PR 댓글로 팀장에게 묻는다.

### 5.7 금지 사항 요약
- `main` 직접 푸시, force push, 직접 머지
- `ollama/ollama`(원본)에 이슈나 PR 생성
- 담당 범위 밖 파일 수정(꼭 필요하면 PR 본문 "범위 밖 변경" 칸에 이유를 적는다)
- `go.mod`, `go.sum` 수정(팀장 제외), `go mod tidy` 실행
- `server/`, `llm/`, `ml/`, `api/` 수정
- 이슈, PR, 커밋, 코드 주석에 이모지 사용
- 토큰, 비밀번호, 개인 경로 같은 민감 정보 커밋

---

## 6. 템플릿 원문 (S-1에서 저장소에 추가)

### 6.1 `.github/ISSUE_TEMPLATE/feature.md`
```markdown
---
name: 기능 작업
about: 기능 추가, 테스트, 문서 등 계획된 작업
title: "[작업ID] type(scope): 요약"
labels: []
assignees: []
---

## 작업 ID
<!-- 예: W-1 (TEAM_GUIDE.md 8장) -->

## 목적
<!-- 이 작업이 왜 필요한지 1~2문장 -->

## 작업 내용
- [ ]
- [ ]

## 수정 / 추가 파일
-

## 완료 조건
- [ ]
- [ ]

## 선행 작업
<!-- 없음 또는 #번호 (작업ID) -->

## 참고
<!-- 관련 함수, 문서 위치 -->
```

### 6.2 `.github/ISSUE_TEMPLATE/bug.md`
```markdown
---
name: 버그 제보
about: 동작 오류, 빌드 실패, 측정값 이상
title: "[BUG] fix(scope): 요약"
labels: ["type:fix"]
assignees: []
---

## 현상

## 재현 방법
1.
2.

## 기대 동작

## 환경
- OS:
- Go 버전:
- Ollama 서버 버전:
- 커밋:

## 로그 / 출력
```

### 6.3 `.github/pull_request_template.md`
```markdown
## 관련 이슈
Closes #

## 작업 ID

## 변경 사항
-

## 변경 파일
-

## 테스트
- 실행한 명령:
- 결과:
- 확인한 OS:

## 체크리스트
- [ ] 이슈 하나만 다룬다
- [ ] 커밋 메시지가 컨벤션을 따른다 (type(scope): 요약, Refs #번호)
- [ ] gofmt, go vet 통과 (Go 변경 시)
- [ ] 관련 테스트 통과
- [ ] 담당 범위 밖 파일을 수정하지 않았다 (수정했다면 아래에 이유 기재)
- [ ] 이모지를 사용하지 않았다

## 범위 밖 변경
<!-- 없음 또는 파일과 이유 -->

## 리뷰어에게
<!-- 특히 봐 줬으면 하는 부분, 설계와 달라진 점 -->
```

### 6.4 `.github/ISSUE_TEMPLATE/config.yml`
```yaml
blank_issues_enabled: true
```

---

## 7. 공통 계약 (모든 담당자가 지켜야 함)

> 라인 번호는 `d63b702` 기준의 참고값이다. 실제 위치는 함수 이름으로 찾는다.
> 실행 흐름: `RunHandler()`(cmd/cmd.go:843) 다음에, 프롬프트가 있으면 `generate()`(:2137), 없으면 `generateInteractive()`(cmd/interactive.go:34)에서 `chat()`(:2010)을 호출한다. 두 함수 모두 `client.Chat()`/`client.Generate()`가 응답을 끝까지 기다리는 구간에 샘플러를 감싼다.

### 7.1 파일 소유권
| 파일 | 담당 | 비고 |
|---|---|---|
| `cmd/profile.go`, `cmd/profile_test.go` | 송하성 | 자료구조, Sampler, summarize |
| `cmd/profile_process.go`, `cmd/profile_process_test.go` | 송하성 | Ollama 프로세스 CPU·RAM 측정 |
| `cmd/gpu_darwin.go`, `cmd/gpu_windows.go`, `cmd/gpu_other.go`, `cmd/gpu_*_test.go` | 송하성 | OS별 GPU |
| `go.mod`, `go.sum`, `.github/`, `project-docs/TEAM_GUIDE.md` | 송하성 | |
| `cmd/profile_print.go`, `cmd/profile_print_test.go` | 우혁 | 출력 형식 |
| `cmd/cmd.go`의 `runOptions`, run 플래그 등록, `RunHandler`, `chat()`, `generate()` | 우혁 | J-4는 정석환이 이 위에 저장 호출만 추가 |
| `cmd/interactive.go` | 우혁 | `/set profile` |
| `cmd/stats.go`, `cmd/stats_test.go` | 정석환 | 기록 저장, 조회, 집계 |
| `cmd/cmd.go`의 `NewCLI()` 안 stats 명령 등록 부분 | 정석환 | |
| `cmd/profile_edge_test.go`, `cmd/stats_edge_test.go` | 김나영 | 경계값 테스트 |
| `tools/profile-bench/**` | 김나영 | 측정, 그래프 스크립트 |
| `project-docs/PROFILE.md`, `project-docs/TEST_REPORT.md`, `project-docs/RESULTS.md`, `project-docs/results/**` | 김나영 | 문서, 결과 |

### 7.2 측정 자료구조와 Sampler (`cmd/profile.go`, H-1에서 확정)
```go
package cmd

// 0.2초마다 1건 수집
type ProfileSample struct {
	At         time.Time
	CPUPercent float64 // Ollama 프로세스 CPU 합 / 논리 코어 수 (0~100)
	RAMUsed    uint64  // Ollama 프로세스 RSS 합 (byte)
	GPUPercent float64 // GPU 사용률(%), 읽지 못하면 -1
	GPUMemUsed uint64  // GPU 메모리 (byte), 읽지 못하면 0
}

// 응답 1건 요약
type ProfileResult struct {
	OS         string  // runtime.GOOS
	Samples    int
	CPUAvg     float64
	CPUPeak    float64
	RAMPeak    uint64
	GPUAvg     float64 // 유효 GPU 샘플이 하나도 없으면 -1
	GPUPeak    float64 // 유효 GPU 샘플이 하나도 없으면 -1
	GPUMemPeak uint64
}

const defaultSampleInterval = 200 * time.Millisecond

type Sampler struct { /* 비공개 필드: interval, read func() ProfileSample, mu, samples, stop, done ... */ }

func newSampler(interval time.Duration) *Sampler                // read = readSample
func (s *Sampler) Start(ctx context.Context)                     // 고루틴 시작. ctx 취소 시에도 종료
func (s *Sampler) Stop() ProfileResult                           // 고루틴 종료를 기다린 뒤 요약 반환. 여러 번 호출해도 안전
func summarize(osName string, samples []ProfileSample) ProfileResult
func readSample() ProfileSample                                  // readOllamaUsage() + readGPU() 조합
```
- `summarize` 규칙: CPU avg/peak는 전체 샘플 기준이다. GPU avg/peak는 `GPUPercent >= 0`인 샘플만 사용하고, 유효 샘플이 없으면 -1이다. 샘플이 0개이면 `Samples=0`, 나머지는 0, GPU 값은 -1이다.

### 7.3 OS별 GPU (`cmd/gpu_*.go`)
```go
// 모든 OS 파일에 같은 시그니처로 존재 (빌드 태그로 하나만 컴파일)
func readGPU() (percent float64, memUsed uint64) // 실패 시 (-1, 0). 100ms 안에 반환
```
| 파일 | 빌드 태그 | 구현 |
|---|---|---|
| `gpu_darwin.go` | `//go:build darwin` | `ioreg -r -d 1 -w 0 -c IOAccelerator`의 `"Device Utilization %"`, `"In use system memory"` |
| `gpu_windows.go` | `//go:build windows` | 기본 (-1, 0). H-4(선택)에서 구현 |
| `gpu_other.go` | `//go:build !darwin && !windows` | 항상 (-1, 0) |

### 7.4 출력 형식 (`ProfileResult.Print`, W-2)
```go
func (r ProfileResult) Print(w io.Writer) // 호출부에서는 os.Stderr 전달 (Summary()와 같은 스트림)
```
정상:
```
-- profile (darwin, 41 samples) ----------------
CPU      avg 38%    peak 71%
RAM      peak 11.2 GB
GPU      avg 92%    peak 100%
GPU mem  peak 2.6 GB
```
GPU를 읽지 못한 경우(GPUAvg == -1):
```
GPU      N/A
GPU mem  N/A
```
샘플 0개:
```
-- profile (windows, 0 samples) ----------------
no samples collected (response finished too quickly)
```
- 퍼센트는 정수로 반올림(`%.0f%%`), 바이트는 GB 소수 첫째 자리(`%.1f GB`, 1 GB = 1024^3 byte)로 표시한다.
- 첫 줄 구분선은 `-`로 채워 전체 48자로 맞춘다.
- `api.Metrics.Summary()`는 수정하지 않는다. 그 호출 바로 다음에 `Print`를 호출한다. `--profile`만 켜고 `--verbose`를 끈 경우에도 Print는 출력된다.

### 7.5 기록 파일 스키마 (`~/.ollama/stats.jsonl`, J-1)
한 줄에 JSON 객체 1개를 append한다(JSON Lines).
```json
{"time":"2026-11-20T14:03:11+09:00","model":"llama3.2:3b","prompt":"양자역학을 설명해줘","os":"darwin","eval_count":198,"eval_rate":24.8,"total_ms":8120,"cpu_avg":38.2,"cpu_peak":71.0,"ram_peak":12025908224,"gpu_avg":92.1,"gpu_peak":100,"gpu_mem_peak":2791728742}
```
```go
type StatsRecord struct {
	Time       time.Time `json:"time"`
	Model      string    `json:"model"`
	Prompt     string    `json:"prompt"`       // 앞 40자(rune), 줄바꿈은 공백으로
	OS         string    `json:"os"`
	EvalCount  int       `json:"eval_count"`
	EvalRate   float64   `json:"eval_rate"`    // EvalCount / EvalDuration(초), EvalDuration=0이면 0
	TotalMs    int64     `json:"total_ms"`     // TotalDuration.Milliseconds()
	CPUAvg     float64   `json:"cpu_avg"`
	CPUPeak    float64   `json:"cpu_peak"`
	RAMPeak    uint64    `json:"ram_peak"`
	GPUAvg     float64   `json:"gpu_avg"`      // -1 = N/A
	GPUPeak    float64   `json:"gpu_peak"`     // -1 = N/A
	GPUMemPeak uint64    `json:"gpu_mem_peak"`
}

func statsFilePath() (string, error)                                         // 환경변수 OLLAMA_STATS_FILE 우선, 없으면 ~/.ollama/stats.jsonl
func newStatsRecord(model, prompt string, m api.Metrics, r ProfileResult) StatsRecord
func appendStats(path string, rec StatsRecord) error                         // 폴더 생성, O_APPEND|O_CREATE|O_WRONLY, 0644
func loadStats(path string) ([]StatsRecord, error)                           // 파일 없으면 빈 슬라이스와 nil. 깨진 줄은 건너뜀
func summarizeByModel(recs []StatsRecord) []modelSummary                     // 모델 이름순 정렬
func StatsHandler(cmd *cobra.Command, args []string) error
```
- `OLLAMA_STATS_FILE`은 테스트와 측정 스크립트용 경로 지정 변수다.

### 7.6 테스트 이름 규칙
`TestProfile*`, `TestSampler*`, `TestSummarize*`, `TestGPU*`, `TestParseIOReg*`, `TestStats*` 중 하나로 시작한다. 실제 기기와 서버가 필요한 테스트는 환경변수 `OLLAMA_PROFILE_LIVE=1`일 때만 실행되게 하고, 아니면 `t.Skip`한다.

### 7.7 PR 전 공통 검사
```bash
gofmt -l cmd/                     # 출력이 없어야 함 (있으면 gofmt -w <파일>)
go vet ./cmd/
go test ./cmd/ -run 'Profile|Sampler|Summarize|GPU|IOReg|Stats' -count=1
go build -o ollama-dev .          # Windows: ollama-dev.exe
```
가능하면 동시성 코드(H-1, H-3, W-3, W-4)는 `-race`도 붙여 실행한다.

---

## 8. 역할별 작업 명세

### 8.0 전체 이슈 맵과 의존 관계
| ID | 담당 | 제목 | 선행 |
|---|---|---|---|
| S-1 | 송하성 | chore(github): 이슈·PR 템플릿과 작업 지시서 추가 | - |
| S-2 | 송하성 | ci(ci): 팀 CI 워크플로 추가 | S-1 |
| H-1 | 송하성 | feat(profile): 측정 자료구조와 샘플러 추가 | S-1 |
| H-2 | 송하성 | feat(gpu): macOS GPU 사용률·메모리 읽기 | H-1 |
| H-3 | 송하성 | feat(profile): Ollama 프로세스만 합산하는 CPU·RAM 측정 | H-1 |
| H-4 | 송하성 | feat(gpu): Windows GPU 측정 (선택) | H-1 |
| W-1 | 우혁 | feat(run): --profile 플래그 추가 | S-1 |
| W-2 | 우혁 | feat(profile): 측정 결과 출력 형식 구현 | H-1 |
| W-3 | 우혁 | feat(run): chat()에 프로파일링 연결 | H-1, W-1, W-2 |
| W-4 | 우혁 | feat(run): generate()에 프로파일링 연결 | W-3 |
| W-5 | 우혁 | feat(interactive): /set profile, /set noprofile 추가 | W-3 |
| J-1 | 정석환 | feat(stats): 측정 기록 구조체와 jsonl 저장·읽기 | H-1 |
| J-2 | 정석환 | feat(stats): ollama stats 명령과 기록 목록 표 | J-1 |
| J-3 | 정석환 | feat(stats): 모델별 평균 표 | J-2 |
| J-4 | 정석환 | feat(stats): --profile 결과 자동 저장 | J-1, W-4 |
| J-5 | 정석환 | feat(stats): stats --clear 기록 초기화 (선택) | J-2 |
| N-1 | 김나영 | test(profile): 요약·출력 경계값 테스트 보강 | H-1, W-2 |
| N-2 | 김나영 | test(stats): 저장·집계 경계값 테스트 보강 | J-3 |
| N-3 | 김나영 | feat(bench): 모델 비교 측정 스크립트 | - (계약 7.5 기준) |
| N-4 | 김나영 | feat(bench): 결과 집계·그래프 스크립트 | - (계약 7.5 기준) |
| N-5 | 김나영 | docs(docs): 기능 사용 설명서 | W-5, J-3 |
| N-6 | 김나영 | docs(docs): OS별 수동 테스트 보고서 | J-4 |
| N-7 | 김나영 | docs(docs): 모델 비교 측정 결과 | N-3, N-4, J-4 |

```
S-1 --> H-1 --+--> H-2, H-3, H-4
              +--> W-2 --+
W-1 ---------------------+--> W-3 --> W-4 --> W-5
              +--> J-1 --> J-2 --> J-3 ; (J-1 + W-4) --> J-4
N-3, N-4 : 즉시 시작 가능
```

---

### 8.1 송하성 (팀장) - 협업 환경, 샘플링 모듈, GPU

#### S-1 `[S-1] chore(github): 이슈·PR 템플릿과 작업 지시서 추가`
- 브랜치: `chore/<번호>-team-templates` / 라벨: `type:chore`, `area:infra`
- 내용: 업스트림 `.github/ISSUE_TEMPLATE/*`를 제거하고 6장의 템플릿 4개를 추가한다. 이 문서를 `project-docs/TEAM_GUIDE.md`로 추가한다.
- 커밋 계획:
  1. `chore(github): 업스트림 이슈 템플릿 제거`
  2. `chore(github): 기능 작업 이슈 템플릿 추가`
  3. `chore(github): 버그 이슈 템플릿 추가`
  4. `chore(github): PR 템플릿 추가`
  5. `docs(docs): 팀 작업 지시서 추가`
- 완료 조건: GitHub에서 New issue를 누르면 "기능 작업"과 "버그 제보"가 보이고, PR 생성 시 템플릿이 자동으로 채워진다.

#### S-2 `[S-2] ci(ci): 팀 CI 워크플로 추가`
- 목적: Windows 팀원 3명의 로컬 빌드 환경이 불안정해도 PR마다 3개 OS에서 빌드와 테스트를 확인한다.
- 파일: `.github/workflows/team-ci.yml`
```yaml
name: team-ci
on:
  pull_request:
    branches: [main]
  push:
    branches: [main]
jobs:
  cmd-test:
    strategy:
      fail-fast: false
      matrix:
        os: [macos-latest, windows-latest, ubuntu-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: gofmt
        if: runner.os != 'Windows'
        run: test -z "$(gofmt -l cmd/)"
      - name: vet
        run: go vet ./cmd/
      - name: test
        run: go test ./cmd/ -run "Profile|Sampler|Summarize|GPU|IOReg|Stats" -count=1
```
- 커밋 계획: 1 `ci(ci): 팀 CI 워크플로 추가` 2 `docs(docs): 지시서에 CI 확인 방법 추가`
- 완료 조건: 이 PR에서 team-ci가 3개 OS 모두 실행된다. 업스트림 워크플로는 실행되지 않는다.

#### H-1 `[H-1] feat(profile): 측정 자료구조와 샘플러 추가`
- 파일: `go.mod`, `go.sum`, `cmd/profile.go`, `cmd/profile_test.go`, `cmd/gpu_darwin.go`, `cmd/gpu_windows.go`, `cmd/gpu_other.go`
- 내용: 7.2 계약을 그대로 구현한다. 이 단계의 `readSample()`은 gopsutil의 시스템 전체 CPU·RAM 값을 쓰고(H-3에서 교체), `readGPU()`는 세 OS 모두 (-1, 0)을 반환하는 기본 구현이다.
- Sampler 요구사항:
  - `read` 함수 필드를 두어 테스트에서 가짜 측정 함수를 주입할 수 있게 한다.
  - 내부 stop 채널과 `sync.WaitGroup`으로 Stop이 고루틴 종료를 기다린다. Stop 중복 호출과 ctx 취소 후 Stop 모두 안전해야 한다.
  - samples 접근은 mutex로 보호한다(`-race` 통과).
- 커밋 계획:
  1. `build(deps): gopsutil v4 의존성 추가`
  2. `feat(profile): ProfileSample, ProfileResult 자료구조 추가`
  3. `feat(profile): 샘플 요약 함수 summarize 추가`
  4. `test(profile): summarize 단위 테스트 추가`
  5. `feat(gpu): OS별 readGPU 기본 구현 추가`
  6. `feat(profile): 0.2초 주기 Sampler Start, Stop 구현`
  7. `test(profile): Sampler 주입 측정 함수 기반 동작 테스트 추가`
- 완료 조건: 7.7 검사 통과, `-race` 통과, 7.2 시그니처와 완전히 일치

#### H-2 `[H-2] feat(gpu): macOS GPU 사용률·메모리 읽기`
- 파일: `cmd/gpu_darwin.go`, `cmd/gpu_darwin_test.go`
- 내용:
  - `parseIOReg(out []byte) (percent float64, mem uint64, ok bool)`: 정규식으로 `"Device Utilization %"=(\d+)`, `"In use system memory"=(\d+)`을 찾는다. 가속기가 여러 개면 사용률은 최댓값, 메모리는 합계로 한다.
  - `readGPU()`: `exec.CommandContext`에 100ms 타임아웃을 주고 ioreg를 실행한 뒤 parse한다. 실패하면 (-1, 0)을 반환한다.
  - 테스트는 실제 ioreg 출력 일부를 문자열 픽스처로 넣어 parse만 검증한다(CI 맥 러너 의존 없음).
- 커밋 계획: 1 `feat(gpu): ioreg 출력 파싱 함수 추가` 2 `test(gpu): ioreg 파싱 픽스처 테스트 추가` 3 `feat(gpu): ioreg 실행과 타임아웃 처리로 readGPU 구현` 4 `test(gpu): 실기기 live 테스트 추가(OLLAMA_PROFILE_LIVE)`
- 완료 조건: 맥북에서 live 테스트를 실행하면 GPU 값이 0 이상이다.

#### H-3 `[H-3] feat(profile): Ollama 프로세스만 합산하는 CPU·RAM 측정`
- 파일: `cmd/profile_process.go`, `cmd/profile_process_test.go`, `cmd/profile.go`(readSample 교체)
- 내용:
  - `isOllamaProcessName(name string) bool`: 대소문자를 무시하고 `ollama` 또는 `ollama.exe`와 **정확히 일치**할 때만 true. `ollama app.exe`(Windows 트레이)와 `ollama-dev`(우리 CLI)는 제외된다.
  - 자기 자신 PID(`os.Getpid()`)는 항상 제외한다.
  - 프로세스 목록은 매 샘플마다 다시 조회한다(요청 후 러너가 새로 뜰 수 있음). CPU 델타 계산을 위해 `map[int32]*process.Process` 캐시를 유지하고, 처음 본 프로세스는 기준값만 잡는다.
  - CPU = 프로세스별 CPU% 합 / `runtime.NumCPU()`, RAM = RSS 합. 개별 프로세스 오류(종료, 권한)는 무시한다.
  - 실행 중 `ps aux | grep -i ollama`(맥), 작업 관리자(Windows)로 러너 프로세스 이름을 확인하고, 이름이 다르면 판별 함수에 추가한 뒤 PR에 기록한다.
  - readSample 1회가 100ms를 넘으면 프로세스 목록 조회를 1초 주기로 바꾸고 캐시를 재사용한다.
- 커밋 계획: 1 `feat(profile): Ollama 프로세스 이름 판별 함수 추가` 2 `test(profile): 프로세스 이름 판별 테스트 추가` 3 `feat(profile): Ollama 프로세스 CPU, RAM 합산 측정 구현` 4 `refactor(profile): readSample을 프로세스별 측정으로 교체` 5 `test(profile): 프로세스 측정 live 테스트 추가`
- 완료 조건: 다른 앱(브라우저 등)을 켜고 꺼도 유휴 상태 CPU 값이 거의 변하지 않는다(PR에 전후 수치 기록).

#### H-4 (선택) `[H-4] feat(gpu): Windows GPU 측정`
- 파일: `cmd/gpu_windows.go`, `cmd/gpu_windows_test.go`
- 내용: 1순위는 `nvidia-smi --query-gpu=utilization.gpu,memory.used --format=csv,noheader,nounits`(PATH에 있을 때)이다. 없으면 PowerShell `Get-Counter '\GPU Engine(*engtype_3D)\Utilization Percentage'`, `'\GPU Adapter Memory(*)\Dedicated Usage'`를 사용한다. PowerShell은 느리므로 백그라운드 고루틴이 1초마다 갱신하고 `readGPU()`는 캐시 값을 즉시 반환한다. 파싱 함수는 픽스처로 테스트한다.
- 검증: Windows 팀원(김나영) 노트북에서 실행한 결과를 PR에 첨부한다.
- 커밋 계획: 1 `feat(gpu): nvidia-smi 출력 파싱 추가` 2 `feat(gpu): 성능 카운터 출력 파싱 추가` 3 `test(gpu): Windows GPU 파싱 테스트 추가` 4 `feat(gpu): 캐시 갱신 방식 Windows readGPU 구현`
- 11/28까지 끝나지 않으면 N/A를 유지하고 "한계"로 문서화한다.

---

### 8.2 우혁 - --profile 옵션과 출력 형식

#### W-1 `[W-1] feat(run): --profile 플래그 추가`
- 파일: `cmd/cmd.go`, `cmd/cmd_profile_test.go`(신규)
- 먼저 확인할 것: `chat()`과 `generate()`가 verbose 값을 어떻게 읽는지 코드에서 확인한다.
  - `cmd.Flags().GetBool("verbose")`처럼 플래그에서 직접 읽는다면 → profile도 같은 방식으로 읽는다. `runOptions.Profile` 필드는 추가하지 않고, 설계 문서와 다른 점을 PR "리뷰어에게"에 적는다.
  - `runOptions`를 통해 전달한다면 → `runOptions`에 `Profile bool`을 추가하고 `RunHandler`에서 `cmd.Flags().GetBool("profile")`로 채운다(keepalive, think와 같은 패턴).
  - 이 판단은 W-5의 `/set profile`이 정상 동작하는 방식(플래그 값 변경)과 맞아야 한다.
- 플래그 등록: run 명령에서 `verbose` 등록 바로 아래에 `Bool("profile", false, "Show CPU/RAM/GPU usage for each response")`를 추가한다.
- 테스트: `NewCLI()`에서 run 명령을 찾아 `Flags().Lookup("profile")`이 존재하고 기본값이 false인지 확인한다.
- 커밋 계획: 1 `feat(run): run 명령에 --profile 플래그 등록` 2 `feat(run): RunHandler에서 profile 옵션 전달` (플래그 직접 읽기 방식이면 생략 가능) 3 `test(run): --profile 플래그 등록 테스트 추가`
- 완료 조건: `./ollama-dev run --help`에 `--profile`이 보인다. 아직 측정 동작은 없다.

#### W-2 `[W-2] feat(profile): 측정 결과 출력 형식 구현`
- 파일: `cmd/profile_print.go`, `cmd/profile_print_test.go`
- 내용: 7.4 계약을 그대로 구현한다. 테스트는 `bytes.Buffer`에 출력한 문자열을 기대값과 비교한다(정상, GPU N/A, 샘플 0개, RAM 0).
- 커밋 계획: 1 `feat(profile): ProfileResult.Print 머리줄과 CPU, RAM 출력` 2 `feat(profile): GPU, GPU 메모리 출력과 N/A 처리` 3 `feat(profile): 샘플이 없을 때 안내 문구 출력` 4 `test(profile): Print 출력 형식 테스트 추가`
- 완료 조건: 출력이 7.4 예시와 글자 단위로 일치한다.

#### W-3 `[W-3] feat(run): chat()에 프로파일링 연결`
- 파일: `cmd/cmd.go`(chat 함수)
- 내용:
  ```go
  var smp *Sampler
  if profile {                       // W-1에서 정한 방식으로 읽은 값
      smp = newSampler(defaultSampleInterval)
      smp.Start(ctx)                 // client.Chat 호출 직전
  }
  err := client.Chat(ctx, req, fn)
  var res ProfileResult
  if smp != nil {
      res = smp.Stop()               // 에러 경로에서도 반드시 Stop (고루틴 누수 방지)
  }
  if err != nil { ...기존 처리... }
  if verbose { latest.Summary() }    // 기존 코드 유지
  if smp != nil { res.Print(os.Stderr) }
  ```
  - Ctrl+C로 취소된 경우에는 Print하지 않는다.
  - 기존 출력 순서와 동작은 `--profile`이 꺼져 있으면 완전히 동일해야 한다.
- 커밋 계획: 1 `feat(run): chat 요청 구간에 샘플러 시작, 종료 추가` 2 `feat(run): chat 응답 후 프로파일 결과 출력` 3 `fix(run): chat 오류, 취소 시 샘플러 정리`
- 완료 조건: 대화형 모드에서 `./ollama-dev run llama3.2:1b --profile` 실행 후 질문하면 응답 아래에 프로파일이 출력된다(PR에 실행 결과 첨부).

#### W-4 `[W-4] feat(run): generate()에 프로파일링 연결`
- 파일: `cmd/cmd.go`(generate 함수)
- 내용: W-3과 같은 패턴을 `client.Generate()` 구간에 적용한다.
- 커밋 계획: 1 `feat(run): generate 요청 구간에 샘플러 시작, 종료 추가` 2 `feat(run): generate 응답 후 프로파일 결과 출력` 3 `fix(run): generate 오류, 취소 시 샘플러 정리`
- 완료 조건: `./ollama-dev run llama3.2:1b --profile "하늘은 왜 파란가?"`와 `--verbose --profile` 조합 모두 정상 출력된다.

#### W-5 `[W-5] feat(interactive): /set profile, /set noprofile 추가`
- 파일: `cmd/interactive.go`
- 내용: `/set verbose`, `/set quiet` 처리(약 319줄) 옆에 `profile`(플래그 true + "Set 'profile' mode." 안내)과 `noprofile`(false + "Set 'noprofile' mode.")을 같은 방식으로 추가한다. `/set` 도움말 목록에도 두 항목을 추가한다.
- 커밋 계획: 1 `feat(interactive): /set profile, /set noprofile 명령 추가` 2 `docs(interactive): /set 도움말에 profile 항목 추가`
- 완료 조건: 대화 중 `/set profile`로 켜고 `/set noprofile`로 끌 수 있다.

---

### 8.3 정석환 - 기록 저장과 ollama stats

#### J-1 `[J-1] feat(stats): 측정 기록 구조체와 jsonl 저장·읽기`
- 파일: `cmd/stats.go`, `cmd/stats_test.go`
- 내용: 7.5 계약 중 `StatsRecord`, `statsFilePath`, `newStatsRecord`, `appendStats`, `loadStats`를 구현한다.
  - `statsFilePath`: `OLLAMA_STATS_FILE`가 있으면 그 값, 없으면 `os.UserHomeDir()` + `.ollama/stats.jsonl`
  - `appendStats`: 상위 폴더가 없으면 생성(`os.MkdirAll`)하고, `json.Marshal` + `\n`으로 한 줄을 쓴다.
  - `loadStats`: `bufio.Scanner`의 버퍼를 1MB로 늘리고, 파싱 실패 줄은 건너뛴다. 파일이 없으면 `(nil, nil)`을 반환한다.
  - `newStatsRecord`: prompt는 줄바꿈을 공백으로 바꾸고 40 rune으로 자른다. Time은 `time.Now()`, OS는 `r.OS`를 쓴다.
  - 테스트는 `t.TempDir()`와 `t.Setenv("OLLAMA_STATS_FILE", ...)`을 사용한다.
- 커밋 계획: 1 `feat(stats): StatsRecord 구조체 정의` 2 `feat(stats): 기록 파일 경로 결정 함수 추가` 3 `feat(stats): jsonl 한 줄 추가 저장 구현` 4 `feat(stats): jsonl 읽기와 손상 줄 건너뛰기 구현` 5 `feat(stats): Metrics와 ProfileResult로 기록 생성` 6 `test(stats): 저장, 읽기 왕복 테스트 추가`
- 완료 조건: 저장한 기록을 다시 읽으면 모든 필드가 같다.

#### J-2 `[J-2] feat(stats): ollama stats 명령과 기록 목록 표`
- 파일: `cmd/stats.go`, `cmd/cmd.go`(NewCLI 등록 부분만)
- 내용:
  - `statsCmd`: `Use: "stats [MODEL]"`, `Short: "Show saved --profile measurements"`, `Args: cobra.MaximumNArgs(1)`, `RunE: StatsHandler`. 서버 연결이 필요 없으므로 서버 확인용 PreRunE(`checkServerHeartbeat`)를 붙이지 않는다.
  - 플래그 `--limit int`(기본 20): 최신 N개만 표시
  - MODEL 인자가 있으면 해당 모델만 표시한다(`llama3.2`처럼 태그 없이 주면 `llama3.2:*` 모두 포함).
  - 표는 `ListRunningHandler`(ollama ps)의 tablewriter 사용 방식을 그대로 따른다. 열: `TIME | MODEL | PROMPT | TOKENS/S | CPU AVG | CPU PEAK | RAM PEAK | GPU AVG | GPU PEAK | GPU MEM`
  - 시간은 `2006-01-02 15:04` 형식, GPU -1은 `N/A`, 바이트는 7.4와 같은 GB 표기
  - 기록이 없으면 `No profile records yet. Run 'ollama run MODEL --profile' first.`를 출력한다.
- 커밋 계획: 1 `feat(stats): stats 명령 등록` 2 `feat(stats): 기록 목록 표 출력` 3 `feat(stats): MODEL 인자로 기록 필터링` 4 `feat(stats): --limit 플래그 추가` 5 `feat(stats): 기록이 없을 때 안내 문구 출력` 6 `test(stats): 필터링, limit 테스트 추가`
- 완료 조건: `OLLAMA_STATS_FILE`로 샘플 파일을 지정하고 `./ollama-dev stats`를 실행하면 표가 출력된다.

#### J-3 `[J-3] feat(stats): 모델별 평균 표`
- 파일: `cmd/stats.go`, `cmd/stats_test.go`
- 내용: `summarizeByModel`을 구현한다. 모델별로 `Runs`를 세고, 각 수치는 실행별 값의 평균이다. GPU 평균은 -1을 제외하고 계산하며, 전부 -1이면 -1이다. 목록 표 아래에 빈 줄을 두고 `MODEL | RUNS | TOKENS/S | CPU AVG | CPU PEAK | RAM PEAK | GPU AVG | GPU PEAK | GPU MEM` 표를 출력한다. 플래그 `--summary`를 주면 요약 표만 출력한다.
- 커밋 계획: 1 `feat(stats): 모델별 집계 함수 추가` 2 `test(stats): 모델별 집계 테스트 추가` 3 `feat(stats): 모델별 평균 표 출력` 4 `feat(stats): --summary 플래그 추가`
- 완료 조건: 모델 2개, 기록 5개 픽스처에서 평균이 손으로 계산한 값과 같다.

#### J-4 `[J-4] feat(stats): --profile 결과 자동 저장`
- 파일: `cmd/cmd.go`(chat, generate에서 Print 다음 줄만)
- 내용: `res.Print(os.Stderr)` 바로 다음에 아래를 추가한다.
  ```go
  if path, err := statsFilePath(); err == nil {
      rec := newStatsRecord(opts.Model, prompt, latest.Metrics, res)
      if err := appendStats(path, rec); err != nil {
          fmt.Fprintf(os.Stderr, "warning: could not save profile stats: %v\n", err)
      }
  }
  ```
  - prompt: chat()에서는 `opts.Messages`의 마지막 user 메시지 내용, generate()에서는 `opts.Prompt`
  - 저장에 실패해도 실행은 실패시키지 않는다. 샘플 0개인 응답은 저장하지 않는다.
- 커밋 계획: 1 `feat(stats): chat 프로파일 결과 기록 저장` 2 `feat(stats): generate 프로파일 결과 기록 저장` 3 `fix(stats): 저장 실패 시 경고만 출력하도록 처리`
- 완료 조건: `--profile` 실행 2회 후 `./ollama-dev stats`에 2줄이 나온다.

#### J-5 (선택) `[J-5] feat(stats): stats --clear 기록 초기화`
- 내용: `--clear`를 주면 기록 파일을 `stats.jsonl.bak`으로 이름을 바꾸고 안내 문구를 출력한다(삭제하지 않음).
- 커밋 계획: 1 `feat(stats): --clear 플래그로 기록 백업 후 초기화` 2 `test(stats): --clear 동작 테스트 추가`

---

### 8.4 김나영 - 테스트, 측정, 문서

#### N-1 `[N-1] test(profile): 요약·출력 경계값 테스트 보강`
- 파일: `cmd/profile_edge_test.go`
- 케이스(table-driven): 샘플 0개, 샘플 1개, GPU 전부 -1, GPU 일부 -1(평균에서 제외되는지), CPU 0% 고정, RAM 매우 큰 값(64 GB), Print 결과의 N/A 줄, 첫 줄 48자 길이
- 커밋 계획: 1 `test(profile): summarize 경계값 테스트 추가` 2 `test(profile): GPU N/A 혼합 샘플 테스트 추가` 3 `test(profile): Print 경계값 출력 테스트 추가`
- 버그를 찾으면 코드를 직접 고치지 말고 버그 이슈(6.2)를 만들어 담당자에게 할당한다. 실패하는 테스트는 `t.Skip("issue #번호")`로 표시한다.

#### N-2 `[N-2] test(stats): 저장·집계 경계값 테스트 보강`
- 파일: `cmd/stats_edge_test.go`
- 케이스: 빈 파일, 없는 파일, 중간에 깨진 줄, 한글과 이모지가 섞인 40자 초과 prompt 자르기(rune 기준), 태그 없는 모델 필터, GPU -1만 있는 모델 평균, limit이 기록 수보다 큰 경우
- 커밋 계획: 1 `test(stats): 손상, 빈 파일 읽기 테스트 추가` 2 `test(stats): prompt 자르기 테스트 추가` 3 `test(stats): 필터, 집계 경계값 테스트 추가`

#### N-3 `[N-3] feat(bench): 모델 비교 측정 스크립트`
- 파일: `tools/profile-bench/bench.py`, `tools/profile-bench/prompts.txt`, `tools/profile-bench/README.md`, `tools/profile-bench/.gitignore`(`results/` 제외)
- 요구사항(Python 3.10+, 표준 라이브러리만 사용):
  ```
  python tools/profile-bench/bench.py --bin ./ollama-dev \
      --models llama3.2:1b llama3.2:3b llama3.1:8b \
      --prompts tools/profile-bench/prompts.txt \
      --repeat 3 --cooldown 60 --out tools/profile-bench/results
  ```
  - 모델마다: 워밍업 1회(`OLLAMA_STATS_FILE=<out>/warmup.jsonl`, 결과에서 제외) → 프롬프트마다 repeat회 실행(`OLLAMA_STATS_FILE=<out>/stats.jsonl`) → `<bin> stop <model>`
  - 각 실행은 새 프로세스로 `[bin, "run", model, "--profile", prompt]`를 실행한다(대화 누적 방지). 타임아웃 300초, stdout/stderr는 `<out>/logs/<model>-<n>.log`에 저장
  - 실행 사이에 `--cooldown`초 대기, 진행 상황은 `[3/27] llama3.2:3b prompt#2 run#1` 형식으로 출력
  - 실행 전 `<bin> list`로 모델 설치 여부를 확인하고, 없으면 `ollama pull <model>` 안내 후 종료
  - Windows와 macOS에서 모두 동작한다(`pathlib`, `subprocess` 사용, 셸 의존 없음).
- prompts.txt: 한 줄에 프롬프트 하나, 짧은 것, 중간, 긴 출력 유도 각 1개 이상(총 3개)
- 커밋 계획: 1 `feat(bench): 측정 스크립트 인자 처리와 모델 확인` 2 `feat(bench): 워밍업과 반복 실행 구현` 3 `feat(bench): 실행 로그 저장과 진행 표시` 4 `feat(bench): 기본 프롬프트 세트 추가` 5 `docs(bench): 측정 스크립트 사용법 추가`
- 계약 7.5 스키마만 알면 되므로 지금 바로 시작할 수 있다. 실제 실행 검증은 J-4 머지 후 한다.

> 참고: Llama 3.2는 1B·3B만 있다. 8B 비교는 `llama3.1:8b`를 쓴다. 8B는 RAM 16GB 이상 기기에서만 측정한다.

#### N-4 `[N-4] feat(bench): 결과 집계·그래프 스크립트`
- 파일: `tools/profile-bench/plot.py`, `tools/profile-bench/sample_stats.jsonl`(테스트용 가짜 데이터 10줄)
- 요구사항:
  ```
  python tools/profile-bench/plot.py --in tools/profile-bench/results/stats.jsonl \
      --order llama3.2:1b llama3.2:3b llama3.1:8b --out project-docs/results
  ```
  - `summary.csv`: model, runs, eval_rate 평균/표준편차, cpu_avg 평균/표준편차, ram_peak_gb 평균, gpu_avg 평균(-1 제외), gpu_mem_peak_gb 평균
  - 그래프 PNG 4개(matplotlib): `eval_rate.png`, `cpu_avg.png`, `ram_peak.png`, `gpu_avg.png`. 모델별 막대와 표준편차 오차 막대, 축 제목과 단위 포함. 그래프 글꼴은 영문만 사용한다(한글 폰트 깨짐 방지).
  - matplotlib이 없으면 CSV만 만들고 `pip install matplotlib` 안내를 출력한다.
- 커밋 계획: 1 `feat(bench): jsonl 읽기와 모델별 집계` 2 `feat(bench): summary.csv 출력` 3 `feat(bench): 지표별 막대 그래프 생성` 4 `test(bench): 샘플 데이터 추가 및 실행 확인`

#### N-5 `[N-5] docs(docs): 기능 사용 설명서`
- 파일: `project-docs/PROFILE.md`
- 목차: 개요 / 빌드 방법(macOS, Windows) / `--profile` 사용법과 출력 해석 / `/set profile` / `ollama stats` 옵션(`--limit`, `--summary`, MODEL) / 기록 파일 위치와 `OLLAMA_STATS_FILE` / 측정 방식(프로세스 합산, OS별 GPU) / 한계(macOS GPU는 시스템 전체 값, 통합 메모리 해석 주의, 첫 실행은 모델 로드 포함)
- 커밋 계획: 1 `docs(docs): PROFILE 사용 설명서 초안` 2 `docs(docs): stats 명령 설명 추가` 3 `docs(docs): 측정 방식과 한계 추가`

#### N-6 `[N-6] docs(docs): OS별 수동 테스트 보고서`
- 파일: `project-docs/TEST_REPORT.md`
- 내용: Windows(본인 노트북)와 macOS(팀장 맥북)에서 아래 체크리스트를 실행하고 결과와 출력 일부를 표로 정리한다.
  1. `run --help`에 `--profile` 표시
  2. `run MODEL --profile "프롬프트"` 출력
  3. `--verbose --profile` 출력 순서
  4. 대화형 `/set profile`, `/set noprofile`
  5. Ctrl+C 취소 시 오류나 멈춤 없음
  6. `stats`, `stats MODEL`, `stats --summary`, `stats --limit 3`
  7. GPU N/A 표시(해당 OS)
  8. 브라우저를 켜고 끈 상태의 CPU 값 비교(프로세스별 측정 확인)
- 실패 항목은 버그 이슈로 만든다.
- 커밋 계획: 1 `docs(docs): 테스트 체크리스트와 Windows 결과` 2 `docs(docs): macOS 결과 추가`

#### N-7 `[N-7] docs(docs): 모델 비교 측정 결과`
- 파일: `project-docs/RESULTS.md`, `project-docs/results/*.png`, `project-docs/results/summary.csv`
- 내용: 팀장 맥북에서 N-3 스크립트로 측정(팀장과 함께 진행) → N-4로 집계 → 측정 조건(기기, OS, Ollama 버전, 모델, 프롬프트, 반복 횟수, 대기 시간), 결과 표, 그래프, 해석 3~5줄, 한계를 정리한다. 발표 자료에 그대로 쓸 수 있는 형태로 만든다.
- 커밋 계획: 1 `docs(docs): 측정 조건 정리` 2 `docs(docs): 모델 비교 결과 표와 그래프 추가` 3 `docs(docs): 결과 해석과 한계 추가`

---

## 9. 일정과 머지 순서

| 날짜 | 송하성 | 우혁 | 정석환 | 김나영 |
|---|---|---|---|---|
| 11/22(일) | 저장소 준비, S-1 | 환경 준비(4장), 이슈 생성 | 환경 준비, 이슈 생성 | 환경 준비, 이슈 생성 |
| 11/23(월) | H-1 머지, S-2, 시작 공지 | W-1 | (H-1 대기) | N-3 |
| 11/24(화) | H-2 | W-2 | J-1 | N-4 |
| 11/25(수) | H-3 | W-3 | J-2 | N-1 |
| 11/26(목) | 리뷰 집중 | W-4 | J-3 | N-5 초안 |
| 11/27(금) | H-4(선택) | W-5 | J-4 | N-2 |
| 11/28(토) | **코드 프리즈 23:59**, 통합 확인 | 버그 수정 | 버그 수정, J-5(선택) | N-6 |
| 11/29(일) | 측정 진행(맥북) | 발표 자료 | 발표 자료 | N-7, N-5 마무리 |
| 11/30(월) | 최종 리허설 | | | |
| 12/1(화) | **최종 발표** | | | |

- 팀장은 PR이 올라오면 12시간 안에 리뷰한다. 머지 순서는 8.0의 의존 관계를 따른다.
- 11/27까지 J-2가 머지되지 않으면 `--profile` 출력 완성을 우선하고 stats 범위를 줄인다(J-3, J-5 제외).
- 코드 프리즈 이후에는 버그 수정(`fix`)과 문서(`docs`)만 머지한다.

---

## 10. 에이전트에게 주는 지시문

### 10.1 시작 프롬프트 (각자 복사해서 사용)
```
너는 공개SW실무 3조 <이름>의 작업 에이전트다. 저장소는 HaSung2/ollama이고 내 로컬 경로는 <경로>다.

1. 저장소의 project-docs/TEAM_GUIDE.md를 처음부터 끝까지 읽어라. 이 문서의 규칙이 너의 기본 습관보다 우선한다.
2. 내 담당은 8.<번호>절의 <작업ID 목록>이다. 7.1 파일 소유권 밖의 파일은 수정하지 마라.
3. 4장 환경 점검을 실행하고 결과를 표로 보여줘. 문제가 있으면 해결 방법을 안내하고 멈춰라.
4. 아직 만들지 않은 내 이슈를 8장 명세대로 모두 생성해라(5.3 규칙, 6.1 템플릿). 선행 작업이 머지되지 않은 이슈에는 status:blocked 라벨을 붙여라.
5. 의존 관계 순서대로 이슈 하나를 골라 진행해라:
   선행 PR 머지 여부 확인 -> main 최신화 -> 브랜치 생성 -> 커밋 계획대로 구현과 커밋 -> 7.7 검사 -> push -> PR 생성(5.4, 6.3).
6. PR을 만들면 멈추고 PR 링크, 커밋 목록, 테스트 결과를 보고해라. 직접 머지하지 마라.
7. 각 단계가 끝날 때마다 무엇을 왜 했는지 2~3줄로 설명해줘. 나는 깃허브를 처음 쓴다.
8. 이슈, PR, 커밋, 주석에 이모지를 쓰지 마라. 도구가 자동으로 붙이는 이모지 서명도 지워라.
9. 계약(7장)과 실제 코드가 맞지 않거나 판단이 필요한 부분이 있으면 임의로 바꾸지 말고 나에게 물어봐라.
```

### 10.2 상황별 지시문
| 상황 | 에이전트에게 할 말 |
|---|---|
| 다음 이슈 진행 | `TEAM_GUIDE 기준으로 내 다음 이슈를 진행해줘. 선행 PR 머지 여부부터 확인해.` |
| 리뷰 반영 | `PR #<번호>의 리뷰 댓글을 gh pr view <번호> --comments로 읽고 반영해줘. 새 커밋으로 추가하고 force push는 하지 마. 반영 내용은 리뷰 댓글에 답글로 남겨.` |
| 충돌 | `PR #<번호>에 main과 충돌이 있어. 5.6 방법(merge)으로 해결해줘. 내 담당이 아닌 코드가 충돌하면 멈추고 알려줘.` |
| CI 실패 | `PR #<번호>의 team-ci 실패 로그를 gh run view로 확인하고 원인과 수정안을 알려줘. 수정은 새 커밋으로.` |
| 머지 후 정리 | `PR #<번호>가 머지됐어. main 최신화하고 로컬 브랜치 정리한 뒤 다음 이슈 상태를 알려줘.` |
| 버그 발견 | `이 문제를 6.2 버그 템플릿으로 이슈 만들고 담당자 <아이디>에게 할당해줘.` |

---

## 11. 자주 생기는 문제

| 증상 | 원인과 해결 |
|---|---|
| 기여도 그래프에 커밋이 안 보임 | 커밋 이메일이 GitHub 계정 이메일과 다름(`git log --format='%ae' -1`로 확인 후 GitHub Settings > Emails에 추가), 아직 main에 머지되지 않음, 반영까지 최대 24시간 걸림 |
| PR이 ollama/ollama로 만들어짐 | 즉시 닫고, `--repo HaSung2/ollama --base main`을 명시해서 다시 만든다. |
| `permission denied` / 403 (push) | 협업자 초대를 수락하지 않았거나 `gh auth status`의 계정이 다르다. |
| Windows `go build`에서 `gcc not found` | TDM-GCC 설치 후 PATH 등록, 새 터미널 열기 |
| macOS 빌드 시 `-I` 관련 CGO 오류 | 저장소 경로에 공백, 대괄호, 한글이 있음. `~/dev/ollama` 같은 경로로 다시 clone |
| `could not connect to ollama app` | 공식 Ollama 앱(서버)이 실행 중이 아님. 앱을 실행하거나 `ollama serve` |
| 프로파일 GPU가 항상 N/A | Windows는 H-4 전까지 정상 동작이다. macOS면 `ioreg -r -d 1 -w 0 -c IOAccelerator | grep "Device Utilization"` 결과를 버그 이슈에 첨부한다. |
| 커밋에 빌드 결과물이 들어감 | 리뷰 전이면 `git rm --cached ollama-dev` 후 `chore: 빌드 결과물 제거` 커밋. 리뷰어에게 알림 |

---

## 부록. gh 명령 모음
```bash
gh issue list --assignee "@me"                    # 내 이슈
gh issue view <번호>
gh pr list --state open                           # 열린 PR
gh pr list --state merged --search "W-1"          # 선행 작업 머지 확인
gh pr view <번호> --comments                      # 리뷰 댓글
gh pr checks <번호>                               # CI 상태
gh run view --log-failed                          # 실패 로그
git log --oneline origin/main -20                 # main 최근 커밋
```
