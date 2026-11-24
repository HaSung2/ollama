#!/usr/bin/env python3
"""Summarize stats.jsonl written by bench.py into a table and bar charts.

Example:
    python tools/profile-bench/plot.py --in tools/profile-bench/results/stats.jsonl \
        --order llama3.2:1b llama3.2:3b llama3.1:8b --out project-docs/results
"""

from __future__ import annotations

import argparse
import csv
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


CSV_COLUMNS = [
    "model", "runs",
    "eval_rate_mean", "eval_rate_std",
    "cpu_avg_mean", "cpu_avg_std",
    "ram_peak_gb_mean",
    "gpu_avg_mean",
    "gpu_mem_peak_gb_mean",
]


def write_csv(summaries: list[dict], path: Path) -> None:
    """summary.csv. A value that could not be measured (GPU N/A) is left empty."""
    with open(path, "w", newline="", encoding="utf-8") as f:
        w = csv.writer(f)
        w.writerow(CSV_COLUMNS)
        for s in summaries:
            row = []
            for k in CSV_COLUMNS:
                v = s[k]
                if k in ("model", "runs"):
                    row.append(v)
                else:
                    row.append("" if v is None else f"{v:.2f}")
            w.writerow(row)


# (file name, title, y-axis label, mean key, std key)
CHARTS = [
    ("eval_rate.png", "Generation speed", "Tokens per second (tokens/s)",
     "eval_rate_mean", "eval_rate_std"),
    ("cpu_avg.png", "Average CPU usage (Ollama processes)", "CPU usage (%)",
     "cpu_avg_mean", "cpu_avg_std"),
    ("ram_peak.png", "Peak RAM usage (Ollama processes)", "RAM (GB)",
     "ram_peak_gb_mean", "ram_peak_gb_std"),
    ("gpu_avg.png", "Average GPU usage", "GPU usage (%)",
     "gpu_avg_mean", "gpu_avg_std"),
]

BAR_COLOR = "#2a78d6"
TEXT_PRIMARY = "#0b0b0b"
TEXT_SECONDARY = "#52514e"
GRID_COLOR = "#e4e3df"


def draw_charts(summaries: list[dict], out: Path) -> list[Path]:
    """One bar chart per metric: a bar per model with a standard deviation error bar.

    Only English text is used so that no Korean font is needed on any OS.
    """
    import matplotlib

    matplotlib.use("Agg")  # no display needed
    import matplotlib.pyplot as plt

    plt.rcParams.update({
        "font.family": "DejaVu Sans",
        "font.size": 11,
        "axes.edgecolor": TEXT_SECONDARY,
        "axes.labelcolor": TEXT_PRIMARY,
        "xtick.color": TEXT_SECONDARY,
        "ytick.color": TEXT_SECONDARY,
    })
    models = [s["model"] for s in summaries]
    written = []
    for name, title, ylabel, mkey, skey in CHARTS:
        means = [s[mkey] for s in summaries]
        stds = [s[skey] for s in summaries]
        fig, ax = plt.subplots(figsize=(7, 4.5), dpi=150)
        x = range(len(models))
        heights = [m if m is not None else 0 for m in means]
        errs = [e if e is not None else 0 for e in stds]
        bars = ax.bar(x, heights, width=0.55, color=BAR_COLOR, yerr=errs, capsize=6,
                      error_kw={"elinewidth": 1.2, "ecolor": TEXT_PRIMARY, "capthick": 1.2})
        top = max([h + e for h, e in zip(heights, errs)] + [0])
        for bar, m, e in zip(bars, means, errs):
            label = "N/A" if m is None else f"{m:.1f}"
            ax.annotate(label, (bar.get_x() + bar.get_width() / 2, (m or 0) + e),
                        xytext=(0, 4), textcoords="offset points", ha="center",
                        va="bottom", color=TEXT_PRIMARY, fontsize=10)
        ax.set_xticks(list(x), models)
        ax.set_xlabel("Model")
        ax.set_ylabel(ylabel)
        ax.set_title(title, loc="left", color=TEXT_PRIMARY, fontsize=13, pad=12)
        ax.set_ylim(0, top * 1.18 if top > 0 else 1)
        ax.yaxis.grid(True, color=GRID_COLOR, linewidth=0.8)
        ax.set_axisbelow(True)
        for side in ("top", "right"):
            ax.spines[side].set_visible(False)
        ax.text(1.0, -0.2, "Error bars: standard deviation across runs",
                transform=ax.transAxes, ha="right", fontsize=8, color=TEXT_SECONDARY)
        fig.tight_layout()
        path = out / name
        fig.savefig(path)
        plt.close(fig)
        written.append(path)
    return written


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

    out = Path(args.out)
    out.mkdir(parents=True, exist_ok=True)
    csv_path = out / "summary.csv"
    write_csv(summaries, csv_path)
    print(f"\nwrote {csv_path}")

    try:
        import matplotlib  # noqa: F401
    except ImportError:
        print("matplotlib is not installed, so only summary.csv was written.")
        print("To draw the charts: pip install matplotlib")
        return 0
    for path in draw_charts(summaries, out):
        print(f"wrote {path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
