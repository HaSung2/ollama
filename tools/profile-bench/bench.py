#!/usr/bin/env python3
"""Model comparison benchmark for `ollama run --profile`.

Runs each model on each prompt several times with a freshly built CLI
(ollama-dev) and lets the CLI append one record per run to a JSON Lines
file (see TEAM_GUIDE.md 7.5). Only the Python standard library is used so
that the script works the same on Windows and macOS.

Example:
    python tools/profile-bench/bench.py --bin ./ollama-dev \
        --models llama3.2:1b llama3.2:3b llama3.1:8b \
        --prompts tools/profile-bench/prompts.txt \
        --repeat 3 --cooldown 60 --out tools/profile-bench/results
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
import time
from pathlib import Path

MIN_PYTHON = (3, 10)
LIST_TIMEOUT_SEC = 60
RUN_TIMEOUT_SEC = 300
STOP_TIMEOUT_SEC = 60


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(
        description="Run `ollama run MODEL --profile PROMPT` repeatedly and "
        "collect the results in a stats.jsonl file."
    )
    p.add_argument(
        "--bin",
        required=True,
        help="path to the CLI built from this repository (e.g. ./ollama-dev)",
    )
    p.add_argument(
        "--models", nargs="+", required=True, help="models to compare, in order"
    )
    p.add_argument(
        "--prompts",
        required=True,
        help="text file with one prompt per line (blank lines and # comments are ignored)",
    )
    p.add_argument(
        "--repeat", type=int, default=3, help="runs per model and prompt (default 3)"
    )
    p.add_argument(
        "--cooldown",
        type=float,
        default=60,
        help="seconds to wait between runs (default 60)",
    )
    p.add_argument(
        "--out",
        default="tools/profile-bench/results",
        help="output directory (default tools/profile-bench/results)",
    )
    args = p.parse_args(argv)
    if args.repeat < 1:
        p.error("--repeat must be 1 or greater")
    if args.cooldown < 0:
        p.error("--cooldown must be 0 or greater")
    return args


def resolve_bin(path: str) -> Path:
    """Return the CLI path. On Windows `./ollama-dev` also finds `ollama-dev.exe`."""
    candidates = [Path(path)]
    if os.name == "nt" and not path.lower().endswith(".exe"):
        candidates.append(Path(path + ".exe"))
    for c in candidates:
        if c.is_file():
            return c.resolve()
    raise SystemExit(
        f"error: CLI not found: {path}\n"
        "Build it first: go build -o ollama-dev .  (Windows: ollama-dev.exe)"
    )


def load_prompts(path: str) -> list[str]:
    try:
        text = Path(path).read_text(encoding="utf-8")
    except OSError as e:
        raise SystemExit(f"error: cannot read prompts file: {e}")
    prompts = [
        line.strip()
        for line in text.splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    ]
    if not prompts:
        raise SystemExit(f"error: no prompts in {path}")
    return prompts


def installed_models(bin_path: Path) -> set[str]:
    """Names in the first column of `<bin> list`."""
    try:
        proc = subprocess.run(
            [str(bin_path), "list"],
            stdin=subprocess.DEVNULL,
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            timeout=LIST_TIMEOUT_SEC,
        )
    except (OSError, subprocess.TimeoutExpired) as e:
        raise SystemExit(f"error: `{bin_path.name} list` failed: {e}")
    if proc.returncode != 0:
        raise SystemExit(
            f"error: `{bin_path.name} list` failed (exit {proc.returncode}). "
            "Is the Ollama app running?\n" + proc.stderr.strip()
        )
    names = set()
    for line in proc.stdout.splitlines()[1:]:
        cols = line.split()
        if cols:
            names.add(cols[0])
    return names


def is_installed(model: str, names: set[str]) -> bool:
    if model in names:
        return True
    # `ollama list` shows "llama3.2:latest" for a model pulled as "llama3.2".
    return ":" not in model and f"{model}:latest" in names


def check_models(bin_path: Path, models: list[str]) -> None:
    names = installed_models(bin_path)
    missing = [m for m in models if not is_installed(m, names)]
    if missing:
        print("error: these models are not installed:", file=sys.stderr)
        for m in missing:
            print(f"  ollama pull {m}", file=sys.stderr)
        print("Pull them with the commands above and run again.", file=sys.stderr)
        raise SystemExit(1)


def run_once(bin_path: Path, model: str, prompt: str, stats_file: Path) -> str:
    """Run one prompt in a new process so that no conversation carries over.

    Returns "ok", "exit <code>", "timeout" or "error: ...".
    """
    env = os.environ.copy()
    env["OLLAMA_STATS_FILE"] = str(stats_file)
    try:
        proc = subprocess.run(
            [str(bin_path), "run", model, "--profile", prompt],
            stdin=subprocess.DEVNULL,  # piped stdin would be prepended to the prompt
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            env=env,
            timeout=RUN_TIMEOUT_SEC,
        )
    except subprocess.TimeoutExpired:
        return "timeout"
    except OSError as e:
        return f"error: {e}"
    return "ok" if proc.returncode == 0 else f"exit {proc.returncode}"


def stop_model(bin_path: Path, model: str) -> None:
    """Unload the model so that the next model starts from the same state."""
    try:
        subprocess.run(
            [str(bin_path), "stop", model],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            timeout=STOP_TIMEOUT_SEC,
        )
    except (OSError, subprocess.TimeoutExpired) as e:
        print(f"warning: `stop {model}` failed: {e}", file=sys.stderr)


def benchmark(bin_path: Path, models: list[str], prompts: list[str],
              repeat: int, cooldown: float, out: Path) -> int:
    """Warm up and measure every model. Returns the number of failed runs."""
    out.mkdir(parents=True, exist_ok=True)
    stats_file = out / "stats.jsonl"
    warmup_file = out / "warmup.jsonl"
    failed = 0
    first = True

    def wait() -> None:
        nonlocal first
        if not first and cooldown > 0:
            time.sleep(cooldown)
        first = False

    for model in models:
        # The first run includes model loading, so it is kept out of the results.
        wait()
        status = run_once(bin_path, model, prompts[0], warmup_file)
        if status != "ok":
            print(f"warning: warmup of {model} failed ({status})", file=sys.stderr)

        for p_idx, prompt in enumerate(prompts, start=1):
            for r_idx in range(1, repeat + 1):
                wait()
                status = run_once(bin_path, model, prompt, stats_file)
                if status != "ok":
                    failed += 1
                    print(f"warning: {model} prompt#{p_idx} run#{r_idx} failed ({status})",
                          file=sys.stderr)

        stop_model(bin_path, model)
    return failed


def main(argv: list[str] | None = None) -> int:
    if sys.version_info < MIN_PYTHON:
        raise SystemExit("error: Python 3.10 or newer is required")
    args = parse_args(argv)
    bin_path = resolve_bin(args.bin)
    prompts = load_prompts(args.prompts)
    check_models(bin_path, args.models)
    out = Path(args.out).resolve()
    failed = benchmark(bin_path, args.models, prompts, args.repeat, args.cooldown, out)
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
