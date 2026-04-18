#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from typing import Any

from PIL import Image, ImageColor


ASSET_EXTENSIONS = {".png", ".svg", ".jpg", ".jpeg", ".webp"}
MARKDOWN_PATTERN = re.compile(r"\(([^)]+\.(?:png|svg|jpg|jpeg|webp))(?:#[^)]+)?\)", re.IGNORECASE)
DEFAULT_BLUE_HEXES = {"#2563eb", "#dbeafe"}
DEFAULT_BLUE_RGB = [ImageColor.getrgb(color) for color in sorted(DEFAULT_BLUE_HEXES)]
DEFAULT_ACCENT_RGB = ImageColor.getrgb("#2563EB")
DEFAULT_SOFT_RGB = ImageColor.getrgb("#DBEAFE")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Audit documentation visual assets for orphaned files and palette bias.")
    parser.add_argument(
        "--base",
        default="assets/go",
        help="Root directory to scan. Defaults to assets/go",
    )
    parser.add_argument(
        "--format",
        choices=("json", "text"),
        default="json",
        help="Output format. Defaults to json",
    )
    return parser.parse_args()


def color_distance(left: tuple[int, int, int], right: tuple[int, int, int]) -> float:
    return ((left[0] - right[0]) ** 2 + (left[1] - right[1]) ** 2 + (left[2] - right[2]) ** 2) ** 0.5


def is_blue_dominant(pixel: tuple[int, int, int]) -> bool:
    r, g, b = pixel
    return b >= g + 8 and b >= r + 20


def matches_default_blue(pixel: tuple[int, int, int]) -> bool:
    r, g, b = pixel
    channel_range = max(pixel) - min(pixel)
    if channel_range < 14:
        return False
    if color_distance(pixel, DEFAULT_ACCENT_RGB) <= 26 and is_blue_dominant(pixel):
        return True
    if color_distance(pixel, DEFAULT_SOFT_RGB) <= 18 and is_blue_dominant(pixel):
        return True
    return False


def iter_assets(base: Path) -> list[Path]:
    return sorted(
        path for path in base.rglob("*") if path.is_file() and path.suffix.lower() in ASSET_EXTENSIONS
    )


def collect_markdown_refs(base: Path) -> set[Path]:
    refs: set[Path] = set()
    for md_path in base.rglob("*.md"):
        text = md_path.read_text(encoding="utf-8")
        for match in MARKDOWN_PATTERN.finditer(text):
            target = (md_path.parent / match.group(1)).resolve()
            refs.add(target)
    return refs


def analyze_png_palette_bias(path: Path) -> dict[str, Any] | None:
    image = Image.open(path).convert("RGB")
    image.thumbnail((200, 200))
    pixels = list(image.getdata())
    if not pixels:
        return None

    matched = 0
    for pixel in pixels:
        if matches_default_blue(pixel):
            matched += 1
    ratio = matched / len(pixels)
    if ratio < 0.12:
        return None
    return {"path": str(path), "ratio": round(ratio, 4), "kind": "png"}


def analyze_svg_palette_bias(path: Path) -> dict[str, Any] | None:
    content = path.read_text(encoding="utf-8").lower()
    hits = {color: content.count(color) for color in DEFAULT_BLUE_HEXES}
    total_hits = sum(hits.values())
    if total_hits == 0:
        return None
    return {"path": str(path), "hits": total_hits, "colors": hits, "kind": "svg"}


def audit_assets(base: Path) -> dict[str, Any]:
    base = base.resolve()
    assets = iter_assets(base)
    refs = collect_markdown_refs(base)

    orphans = [{"path": str(path), "kind": path.suffix.lower().lstrip(".")} for path in assets if path.resolve() not in refs]

    default_blue_palette_bias: list[dict[str, Any]] = []
    for asset in assets:
        if asset.suffix.lower() == ".png":
            result = analyze_png_palette_bias(asset)
        elif asset.suffix.lower() == ".svg":
            result = analyze_svg_palette_bias(asset)
        else:
            result = None
        if result is not None:
            default_blue_palette_bias.append(result)

    return {
        "base": str(base),
        "asset_count": len(assets),
        "referenced_count": len(refs),
        "orphans": orphans,
        "default_blue_palette_bias": default_blue_palette_bias,
    }


def print_text(report: dict[str, Any]) -> None:
    print(f"Base: {report['base']}")
    print(f"Assets: {report['asset_count']}")
    print(f"Referenced assets: {report['referenced_count']}")
    print(f"Orphans: {len(report['orphans'])}")
    for item in report["orphans"]:
        print(f"  - orphan: {item['path']}")
    print(f"Default-blue palette bias: {len(report['default_blue_palette_bias'])}")
    for item in report["default_blue_palette_bias"]:
        metric = item.get("ratio", item.get("hits"))
        print(f"  - palette-bias: {item['path']} ({metric})")


def main() -> None:
    args = parse_args()
    report = audit_assets(Path(args.base))
    if args.format == "text":
        print_text(report)
        return
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
