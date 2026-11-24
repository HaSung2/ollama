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
from pathlib import Path

MIN_PYTHON = (3, 10)
LIST_TIMEOUT_SEC = 60


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


def main(argv: list[str] | None = None) -> int:
    if sys.version_info < MIN_PYTHON:
        raise SystemExit("error: Python 3.10 or newer is required")
    args = parse_args(argv)
    bin_path = resolve_bin(args.bin)
    prompts = load_prompts(args.prompts)
    check_models(bin_path, args.models)
    print(
        f"{len(args.models)} models, {len(prompts)} prompts, "
        f"{args.repeat} runs each: all models are installed"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
