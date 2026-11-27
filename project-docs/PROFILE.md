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

`--profile`로 측정한 결과는 실행할 때마다 자동으로 기록 파일에 한 줄씩 쌓인다(샘플이 0개인 응답은 저장하지 않는다). `ollama stats`는 이 기록을 표로 보여준다. 서버에 연결하지 않으므로 Ollama 앱이 꺼져 있어도 실행된다.

```bash
./ollama-dev stats                  # 최근 기록 목록 + 모델별 평균
./ollama-dev stats llama3.2         # llama3.2의 모든 태그(llama3.2:1b, llama3.2:3b 등)만
./ollama-dev stats llama3.2:3b      # 해당 태그만
./ollama-dev stats --limit 3        # 최근 3개만
./ollama-dev stats --summary        # 모델별 평균 표만
```

| 옵션 | 기본값 | 설명 |
|---|---|---|
| `MODEL` | 없음 | 해당 모델만 표시한다. 태그 없이 주면(`llama3.2`) 그 모델의 모든 태그를 포함한다. |
| `--limit N` | 20 | 기록 목록에서 최신 N개만 표시한다. |
| `--summary` | 끔 | 기록 목록 없이 모델별 평균 표만 출력한다. |

### 기록 목록 표

| 열 | 의미 |
|---|---|
| `TIME` | 실행 시각(`2006-01-02 15:04` 형식) |
| `MODEL` | 모델 이름 |
| `PROMPT` | 프롬프트 앞 40자(줄바꿈은 공백으로 바뀜) |
| `TOKENS/S` | 생성 속도 = 생성 토큰 수 / 생성 시간(초) |
| `CPU AVG`, `CPU PEAK` | CPU 평균, 최댓값(%) |
| `RAM PEAK` | RAM 최댓값(GB) |
| `GPU AVG`, `GPU PEAK` | GPU 평균, 최댓값(%). 읽지 못했으면 `N/A` |
| `GPU MEM` | GPU 메모리 최댓값(GB) |

### 모델별 평균 표

기록 목록 아래에 빈 줄을 두고 모델별 평균 표가 나온다. 열은 `MODEL | RUNS | TOKENS/S | CPU AVG | CPU PEAK | RAM PEAK | GPU AVG | GPU PEAK | GPU MEM`이다. `RUNS`는 기록 수이고, 나머지는 실행별 값의 평균이다. GPU 평균은 `N/A`인 실행을 빼고 계산하며, 모든 실행이 `N/A`면 `N/A`로 표시한다. 모델은 이름순으로 정렬된다.

기록이 하나도 없으면 다음 안내가 나온다.

```
No profile records yet. Run 'ollama run MODEL --profile' first.
```

## 6. 기록 파일 위치와 `OLLAMA_STATS_FILE`

| OS | 기본 위치 |
|---|---|
| macOS | `~/.ollama/stats.jsonl` |
| Windows | `C:\Users\<사용자>\.ollama\stats.jsonl` |

환경변수 `OLLAMA_STATS_FILE`을 지정하면 기본 위치 대신 그 파일을 쓴다. 측정 스크립트(`tools/profile-bench/bench.py`)가 결과를 따로 모을 때, 테스트에서 실제 기록을 건드리지 않을 때 쓴다.

```bash
# macOS
OLLAMA_STATS_FILE=/tmp/test.jsonl ./ollama-dev stats
```
```powershell
# Windows PowerShell
$env:OLLAMA_STATS_FILE = "C:\temp\test.jsonl"; .\ollama-dev.exe stats
```

파일은 JSON Lines 형식으로, 한 줄이 실행 1회다.

```json
{"time":"2026-11-20T14:03:11+09:00","model":"llama3.2:3b","prompt":"양자역학을 설명해줘","os":"darwin","eval_count":198,"eval_rate":24.8,"total_ms":8120,"cpu_avg":38.2,"cpu_peak":71.0,"ram_peak":12025908224,"gpu_avg":92.1,"gpu_peak":100,"gpu_mem_peak":2791728742}
```

`ram_peak`, `gpu_mem_peak`는 byte 단위이고, GPU를 읽지 못하면 `gpu_avg`, `gpu_peak`가 -1이다. 깨진 줄은 읽을 때 건너뛰므로 파일 일부가 손상되어도 나머지 기록은 그대로 보인다. 기록을 지우고 싶으면 이 파일을 삭제하거나 다른 이름으로 옮긴다.

## 7. 측정 방식

### 언제 측정하나

CLI가 서버에 요청을 보내기 직전에 측정을 시작하고, 응답을 끝까지 받은 직후에 멈춘다. 그 사이에 0.2초마다 샘플 하나를 모은 뒤, 응답이 끝나면 평균과 최댓값을 계산한다. 측정은 별도 고루틴에서 돌아가므로 응답 속도에는 거의 영향을 주지 않는다.

### CPU, RAM: Ollama 프로세스만 합산

컴퓨터 전체 사용량이 아니라 Ollama 프로세스의 사용량만 더한다. 그래서 브라우저 같은 다른 프로그램을 켜고 꺼도 값이 거의 변하지 않는다.

- 대상: 이름이 정확히 `ollama` 또는 `ollama.exe`인 프로세스(대소문자 무시). 서버와, 모델을 실제로 돌리는 러너 프로세스가 여기에 해당한다.
- 제외: Windows 트레이 앱(`ollama app.exe`), 측정을 실행한 CLI 자신(`ollama-dev`)
- 프로세스 목록은 샘플마다 다시 확인한다. 요청 도중에 러너가 새로 뜰 수 있기 때문이다. 다만 측정 1회가 100ms를 넘을 만큼 느리면 목록 확인을 1초에 한 번으로 줄이고, 그 사이에는 이미 찾은 프로세스를 다시 쓴다.
- CPU = 프로세스별 CPU 사용률의 합 / 논리 코어 수, RAM = 프로세스별 RSS(실제 점유 메모리)의 합

### GPU: OS마다 다름

| OS | 방법 | 결과 |
|---|---|---|
| macOS | `ioreg -r -d 1 -w 0 -c IOAccelerator`의 `Device Utilization %`, `In use system memory`(가속기가 여러 개면 사용률은 최댓값, 메모리는 합계). 한 번 읽는 데 100ms가 넘으면 그 샘플은 N/A로 처리 | 사용률과 메모리 표시 |
| Windows | 미구현(선택 작업 H-4) | `N/A` |
| 그 밖의 OS | 미지원 | `N/A` |

GPU 평균과 최댓값은 GPU를 읽은 샘플만으로 계산한다. 읽은 샘플이 하나도 없으면 `N/A`다.

## 8. 한계

- **macOS GPU 값은 시스템 전체 값이다.** `ioreg`는 프로세스별 GPU 사용량을 주지 않는다. 측정 중에 영상 재생이나 화면 효과처럼 GPU를 쓰는 작업이 있으면 값이 함께 올라간다. 측정할 때는 다른 앱을 닫는다.
- **통합 메모리(Apple Silicon) 해석에 주의한다.** CPU와 GPU가 같은 메모리를 나눠 쓰므로 `RAM peak`와 `GPU mem peak`에 같은 메모리가 겹쳐 잡힐 수 있다. 두 값을 더해서 총 사용량으로 해석하면 안 된다.
- **첫 실행에는 모델 로드가 포함된다.** 모델이 메모리에 없으면 첫 요청에서 모델을 불러오는 동안 CPU와 RAM이 크게 오르고 생성 속도는 느리게 나온다. 비교할 때는 첫 실행을 버리거나(워밍업), `tools/profile-bench/bench.py`처럼 워밍업 후 측정한다.
- **짧은 응답은 샘플이 적다.** 0.2초 간격이므로 1초 안에 끝나는 응답은 샘플이 5개 이하라 평균이 불안정하다. 비교에는 길이가 있는 프롬프트를 쓰고 여러 번 반복한다.
- **Windows에서는 GPU를 측정하지 않는다.** 선택 작업(H-4)이 끝나기 전까지 Windows의 GPU 값은 항상 `N/A`다.
- **측정 자체도 약간의 자원을 쓴다.** 0.2초마다 프로세스 목록을 조회하는 비용이 CLI 쪽에 생긴다. CLI는 측정 대상에서 빠지므로 결과 값에는 포함되지 않는다.
