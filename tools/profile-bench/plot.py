#!/usr/bin/env python3
"""Summarize stats.jsonl written by bench.py into a table and bar charts.

Example:
    python tools/profile-bench/plot.py --in tools/profile-bench/results/stats.jsonl \
        --order llama3.2:1b llama3.2:3b llama3.1:8b --out project-docs/results
"""

from __future__ import annotations

import argparse
import json
import statistics
import sys
from pathlib import Path

GB = 1024 ** 3


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(
        description="Aggregate --profile records per model and draw bar charts."
    )
    p.add_argument("--in", dest="inp", required=True, help="stats.jsonl to read")
    p.add_argument(
        "--order",
        nargs="+",
        default=[],
        help="model order in the table and charts (other models follow by name)",
    )
    p.add_argument("--out", required=True, help="output directory")
    return p.parse_args(argv)


def load_records(path: str) -> list[dict]:
    """Read JSON Lines. Blank and broken lines are skipped with a warning."""
    try:
        lines = Path(path).read_text(encoding="utf-8").splitlines()
    except OSError as e:
        raise SystemExit(f"error: cannot read {path}: {e}")
    records, broken = [], 0
    for line in lines:
        if not line.strip():
            continue
        try:
            rec = json.loads(line)
        except json.JSONDecodeError:
            broken += 1
            continue
        if isinstance(rec, dict) and rec.get("model"):
            records.append(rec)
        else:
            broken += 1
    if broken:
        print(f"warning: skipped {broken} broken line(s) in {path}", file=sys.stderr)
    if not records:
        raise SystemExit(f"error: no records in {path}")
    return records


def mean_std(values: list[float]) -> tuple[float | None, float | None]:
    """Mean and sample standard deviation. None when there is no value."""
    if not values:
        return None, None
    std = statistics.stdev(values) if len(values) > 1 else 0.0
    return statistics.fmean(values), std


def num(rec: dict, key: str) -> float:
    try:
        return float(rec.get(key, 0) or 0)
    except (TypeError, ValueError):
        return 0.0


def summarize_model(model: str, recs: list[dict]) -> dict:
    # GPU values are -1 when they could not be read (TEAM_GUIDE 7.5).
    # gpu_mem_peak is 0 in that case, so it is averaged over the same runs.
    gpu_recs = [r for r in recs if num(r, "gpu_avg") >= 0]
    s = {"model": model, "runs": len(recs)}
    s["eval_rate_mean"], s["eval_rate_std"] = mean_std([num(r, "eval_rate") for r in recs])
    s["cpu_avg_mean"], s["cpu_avg_std"] = mean_std([num(r, "cpu_avg") for r in recs])
    s["ram_peak_gb_mean"], s["ram_peak_gb_std"] = mean_std([num(r, "ram_peak") / GB for r in recs])
    s["gpu_avg_mean"], s["gpu_avg_std"] = mean_std([num(r, "gpu_avg") for r in gpu_recs])
    s["gpu_mem_peak_gb_mean"], _ = mean_std([num(r, "gpu_mem_peak") / GB for r in gpu_recs])
    return s


def aggregate(records: list[dict], order: list[str]) -> list[dict]:
    """One summary per model: --order first, then the rest by name."""
    by_model: dict[str, list[dict]] = {}
    for r in records:
        by_model.setdefault(r["model"], []).append(r)
    missing = [m for m in order if m not in by_model]
    if missing:
        print(f"warning: no records for {', '.join(missing)}", file=sys.stderr)
    names = [m for m in dict.fromkeys(order) if m in by_model]
    names += sorted(m for m in by_model if m not in names)
    return [summarize_model(m, by_model[m]) for m in names]


def fmt(v: float | None) -> str:
    return "N/A" if v is None else f"{v:.2f}"


def print_table(summaries: list[dict]) -> None:
    cols = [("MODEL", "model"), ("RUNS", "runs"), ("TOKENS/S", "eval_rate_mean"),
            ("CPU AVG %", "cpu_avg_mean"), ("RAM PEAK GB", "ram_peak_gb_mean"),
            ("GPU AVG %", "gpu_avg_mean"), ("GPU MEM GB", "gpu_mem_peak_gb_mean")]
    rows = [[h for h, _ in cols]]
    for s in summaries:
        rows.append([str(s[k]) if k in ("model", "runs") else fmt(s[k]) for _, k in cols])
    widths = [max(len(r[i]) for r in rows) for i in range(len(cols))]
    for r in rows:
        print("  ".join(c.ljust(w) for c, w in zip(r, widths)).rstrip())


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    records = load_records(args.inp)
    summaries = aggregate(records, args.order)
    print_table(summaries)
    return 0


if __name__ == "__main__":
    sys.exit(main())
