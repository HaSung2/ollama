# profile-bench

`ollama run MODEL --profile`을 모델과 프롬프트마다 여러 번 실행해서 모델별 속도와 자원 사용량을 비교하는 측정 스크립트다. Python 3.10 이상의 표준 라이브러리만 사용하므로 별도 설치 없이 Windows와 macOS에서 실행된다.

## 준비

1. 공식 Ollama 앱(서버)을 실행한다.
2. 저장소 루트에서 CLI를 빌드한다.
   ```bash
   go build -o ollama-dev .        # Windows: go build -o ollama-dev.exe .
   ```
3. 측정할 모델을 받아 둔다. 스크립트는 실행 전에 `ollama-dev list`로 설치 여부를 확인하고, 없으면 `ollama pull` 명령을 안내한 뒤 종료한다.
   ```bash
   ollama pull llama3.2:1b
   ollama pull llama3.2:3b
   ollama pull llama3.1:8b         # RAM 16GB 이상 기기에서만
   ```
   Llama 3.2는 1B와 3B만 있으므로 8B 비교에는 `llama3.1:8b`를 쓴다.

## 실행

저장소 루트에서 실행한다.

```bash
python tools/profile-bench/bench.py --bin ./ollama-dev \
    --models llama3.2:1b llama3.2:3b llama3.1:8b \
    --prompts tools/profile-bench/prompts.txt \
    --repeat 3 --cooldown 60 --out tools/profile-bench/results
```

Windows PowerShell에서는 줄 끝의 `\` 대신 한 줄로 입력한다. `--bin ./ollama-dev`만 줘도 `ollama-dev.exe`를 찾는다.

| 옵션 | 기본값 | 설명 |
|---|---|---|
| `--bin` | (필수) | 이 저장소에서 빌드한 CLI 경로 |
| `--models` | (필수) | 비교할 모델. 적은 순서대로 측정한다 |
| `--prompts` | (필수) | 한 줄에 프롬프트 하나. 빈 줄과 `#` 줄은 무시 |
| `--repeat` | 3 | 모델과 프롬프트 조합마다 반복 횟수 |
| `--cooldown` | 60 | 실행 사이 대기 시간(초). 발열과 이전 실행의 영향을 줄인다 |
| `--out` | `tools/profile-bench/results` | 결과 폴더 |

모델 3개, 프롬프트 3개, 반복 3회면 측정 27회에 워밍업 3회가 더해진다. 대기 60초 기준으로 대기만 30분 정도 걸리므로 충전기를 연결하고 다른 프로그램을 닫은 상태에서 실행한다.

## 동작 방식

모델마다 다음 순서로 진행한다.

1. 워밍업 1회: 첫 번째 프롬프트로 실행해 모델을 메모리에 올린다. 모델 로드 시간이 섞이므로 결과는 `warmup.jsonl`에 따로 저장하고 비교에서 제외한다.
2. 측정: 프롬프트마다 `--repeat`회 실행한다. 결과는 `stats.jsonl`에 쌓인다.
3. 정리: `ollama-dev stop MODEL`로 모델을 내려서 다음 모델이 같은 조건에서 시작하게 한다.

각 실행은 `ollama-dev run MODEL --profile "프롬프트"`를 새 프로세스로 띄운다. 그래서 이전 대화가 누적되지 않는다. 기록은 CLI가 `OLLAMA_STATS_FILE` 환경변수가 가리키는 파일에 직접 저장한다(TEAM_GUIDE 7.5). 한 실행이 300초를 넘으면 중단하고 실패로 센다.

진행 상황은 다음처럼 출력된다.

```
[warmup] llama3.2:3b ... ok (12.4s)
[13/27] llama3.2:3b prompt#2 run#1 ... ok (8.1s)
```

## 결과 폴더

```
results/
  stats.jsonl        측정 기록 (한 줄 = 실행 1회, TEAM_GUIDE 7.5 스키마)
  warmup.jsonl       워밍업 기록 (비교에서 제외)
  logs/
    llama3.2_3b-1.log      실행별 응답과 프로파일 출력
    llama3.2_3b-warmup.log
```

로그 파일 이름은 Windows에서 쓸 수 없는 `:`를 `_`로 바꾼다. `results/`는 `.gitignore`에 등록되어 있어 커밋되지 않는다. 집계와 그래프는 `plot.py`(N-4)로 만든다.

## 문제 해결

| 증상 | 해결 |
|---|---|
| `CLI not found` | 저장소 루트에서 실행했는지, `go build`를 했는지 확인 |
| `list failed` | 공식 Ollama 앱이 실행 중인지 확인 |
| `these models are not installed` | 안내된 `ollama pull` 명령 실행 |
| 실행이 `exit 1`로 실패 | 해당 로그 파일을 열어 오류 확인. `--profile`을 모르는 CLI면 최신 main으로 다시 빌드 |
| `stats.jsonl`이 비어 있음 | 측정 결과 자동 저장(J-4)이 들어간 CLI인지 확인 |
