# --profile, ollama stats 사용 설명서

공개SW실무 3조가 Ollama CLI에 추가한 실행 프로파일링 기능의 사용법이다. 응답을 만드는 동안 Ollama가 CPU, RAM, GPU를 얼마나 쓰는지 측정하고, 측정 결과를 기록해 나중에 모델끼리 비교할 수 있다.

## 목차

1. 개요
2. 빌드 방법(macOS, Windows)
3. `--profile` 사용법과 출력 해석
4. 대화 중 켜고 끄기: `/set profile`
5. `ollama stats` 옵션
6. 기록 파일 위치와 `OLLAMA_STATS_FILE`
7. 측정 방식
8. 한계

## 1. 개요

| 기능 | 설명 |
|---|---|
| `ollama run MODEL --profile` | 응답을 생성하는 동안 0.2초마다 Ollama 프로세스의 CPU, RAM 사용량과 GPU 사용량을 측정하고, 응답이 끝나면 평균값과 최댓값을 출력한다. |
| `ollama stats [MODEL]` | `--profile`로 측정한 결과를 기록 파일에서 읽어 목록과 모델별 평균을 표로 보여준다. |

서버는 공식 설치본을 그대로 쓰고, 이 기능이 들어간 CLI만 따로 빌드해서 쓴다. 아래 예시의 `ollama-dev`가 그 CLI다.

## 2. 빌드 방법

### 공통 준비

1. 공식 Ollama 앱(https://ollama.com/download)을 설치하고 실행한다. 이 앱이 서버 역할을 한다.
2. Go를 설치한다(https://go.dev/dl, 저장소 `go.mod`의 `go` 줄 버전 이상).
3. 저장소를 받는다. 경로에 공백, 대괄호, 한글이 있으면 빌드가 실패하므로 `~/dev/ollama`, `C:\dev\ollama` 같은 경로를 쓴다.
   ```bash
   git clone https://github.com/HaSung2/ollama.git
   cd ollama
   ```

### macOS

```bash
xcode-select --install          # C 컴파일러(Clang), 처음 한 번만
go build -o ollama-dev .
./ollama-dev --version
```

### Windows

Ollama에는 C 코드가 포함되어 있어 Windows에서도 C 컴파일러가 필요하다.

1. TDM-GCC를 설치하고 설치 폴더의 `bin`을 PATH에 등록한다.
2. 새 터미널을 열고 `gcc --version`으로 확인한다.
3. 빌드한다.
   ```powershell
   go build -o ollama-dev.exe .
   .\ollama-dev.exe --version
   ```

`gcc not found` 오류가 나면 PATH 등록 후 터미널을 새로 열었는지 확인한다.

## 3. `--profile` 사용법과 출력 해석

### 실행

```bash
./ollama-dev run llama3.2:1b --profile "하늘은 왜 파란가?"
```

응답이 끝나면 아래와 같은 표가 출력된다(macOS 예시).

```
-- profile (darwin, 41 samples) ----------------
CPU      avg 38%    peak 71%
RAM      peak 11.2 GB
GPU      avg 92%    peak 100%
GPU mem  peak 2.6 GB
```

| 줄 | 의미 |
|---|---|
| 첫 줄 | 측정한 OS와 샘플 수. 샘플은 0.2초마다 1개이므로 41 samples는 약 8초 동안 측정했다는 뜻이다. |
| `CPU` | Ollama 프로세스들의 CPU 사용률 합을 논리 코어 수로 나눈 값(0~100%). `avg`는 측정 구간 평균, `peak`는 최댓값이다. |
| `RAM` | Ollama 프로세스들이 사용한 메모리(RSS) 합의 최댓값. 1 GB = 1024^3 byte다. |
| `GPU` | GPU 사용률의 평균과 최댓값. |
| `GPU mem` | GPU가 사용한 메모리의 최댓값. |

### `--verbose`와 함께 쓰기

```bash
./ollama-dev run llama3.2:1b --verbose --profile "하늘은 왜 파란가?"
```

기존 `--verbose` 통계(토큰 수, 생성 속도 등)가 먼저 나오고, 바로 아래에 프로파일 표가 나온다. 두 출력 모두 표준 에러(stderr)로 나가므로 응답 본문만 파일로 저장해도 섞이지 않는다. `--verbose` 없이 `--profile`만 켜도 프로파일 표는 출력된다.

### 특별한 출력

GPU를 읽지 못한 경우(예: Windows):

```
GPU      N/A
GPU mem  N/A
```

응답이 너무 빨리 끝나 샘플을 하나도 모으지 못한 경우:

```
-- profile (windows, 0 samples) ----------------
no samples collected (response finished too quickly)
```

Ctrl+C로 응답을 취소하면 프로파일 표는 출력하지 않는다.

## 4. 대화 중 켜고 끄기: `/set profile`

프롬프트 없이 실행하면 대화형 모드가 된다. 대화 중에 측정을 켜고 끌 수 있다.

```
./ollama-dev run llama3.2:1b
>>> /set profile
Set 'profile' mode.
>>> 하늘은 왜 파란가?
(응답)
-- profile (darwin, 37 samples) ----------------
...
>>> /set noprofile
Set 'noprofile' mode.
```

`/set` 도움말에도 두 항목이 표시된다. `--profile`로 시작한 대화에서 `/set noprofile`로 끌 수도 있다.

## 5. `ollama stats` 옵션

## 6. 기록 파일 위치와 `OLLAMA_STATS_FILE`

## 7. 측정 방식

## 8. 한계
