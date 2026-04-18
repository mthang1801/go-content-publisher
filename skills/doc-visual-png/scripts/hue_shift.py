#!/usr/bin/env python3
"""Hue-shift existing PNG assets to match folder-specific Visual DNA palettes.

Usage:
  python3 hue_shift.py --folder concurrency --preview
  python3 hue_shift.py --folder gin --preview
  python3 hue_shift.py --folder gin --scope gin/basics --preview
  python3 hue_shift.py --folder advanced --apply
  python3 hue_shift.py --folder all --apply
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path

from PIL import Image, ImageColor


# Visual DNA palettes from design-md skill
FOLDER_PALETTES = {
    "concurrency": {
        "name": "Parallel Movement",
        "accent": "#00BFA5",       # Electric Teal
        "accent_soft": "#E0F7F3",  # Teal mist
        "background": "#1E2A3A",   # Deep Slate (dark mode feel)
        # For light-bg PNGs, shift accent from default blue to teal
        "bg_light": "#F0F8F7",     # Light teal wash
        "signal": "#F59E0B",       # Signal Amber
    },
    "advanced": {
        "name": "Deep Expertise",
        "accent": "#7C3AED",       # Violet expertise  
        "accent_soft": "#EDE9FE",  # Lavender mist
        "background": "#1A1625",   # Deep purple-black
        "bg_light": "#F5F3FF",     # Light violet wash
        "signal": "#F59E0B",       # Amber accent
    },
    "gin": {
        "name": "Runtime Garden",
        "accent": "#15803D",
        "accent_soft": "#DCFCE7",
        "background": "#F2F8F4",
        "bg_light": "#F2F8F4",
        "signal": "#0F766E",
    },
    "fiber": {
        "name": "Edge Velocity",
        "accent": "#EA580C",
        "accent_soft": "#FFEDD5",
        "background": "#FFF7ED",
        "bg_light": "#FFF7ED",
        "signal": "#0F766E",
    },
    "cli": {
        "name": "Terminal Signal",
        "accent": "#65A30D",
        "accent_soft": "#ECFCCB",
        "background": "#F7FEE7",
        "bg_light": "#F7FEE7",
        "signal": "#1F2937",
    },
    "cloud-infra": {
        "name": "Ops Watch",
        "accent": "#BE123C",
        "accent_soft": "#FFE4E6",
        "background": "#FFF1F2",
        "bg_light": "#FFF1F2",
        "signal": "#F59E0B",
    },
    "design-patterns": {
        "name": "Blueprint Forge",
        "accent": "#D97706",
        "accent_soft": "#FFEDD5",
        "background": "#FFFBEB",
        "bg_light": "#FFFBEB",
        "signal": "#1D4ED8",
    },
    "docs": {
        "name": "Reference Ledger",
        "accent": "#0F766E",
        "accent_soft": "#CCFBF1",
        "background": "#F0FDFA",
        "bg_light": "#F0FDFA",
        "signal": "#334155",
    },
    "idioms": {
        "name": "Go Craft",
        "accent": "#9A3412",
        "accent_soft": "#FFEDD5",
        "background": "#FFF7ED",
        "bg_light": "#FFF7ED",
        "signal": "#1D4ED8",
    },
    "export": {
        "name": "Delivery Ledger",
        "accent": "#0F766E",
        "accent_soft": "#CCFBF1",
        "background": "#F0FDFA",
        "bg_light": "#F0FDFA",
        "signal": "#9A3412",
    },
    "fundamental": {
        "name": "First Principles",
        "accent": "#4F46E5",
        "accent_soft": "#E0E7FF",
        "background": "#F8FAFF",
        "bg_light": "#F8FAFF",
        "signal": "#15803D",
    },
    "deployment": {
        "name": "Release Guard",
        "accent": "#B45309",
        "accent_soft": "#FEF3C7",
        "background": "#FFFBEB",
        "bg_light": "#FFFBEB",
        "signal": "#1F2937",
    },
    "observability": {
        "name": "Signal Radar",
        "accent": "#DC2626",
        "accent_soft": "#FEE2E2",
        "background": "#FEF2F2",
        "bg_light": "#FEF2F2",
        "signal": "#0F766E",
    },
    "orm": {
        "name": "Data Rail",
        "accent": "#0F766E",
        "accent_soft": "#CCFBF1",
        "background": "#F0FDFA",
        "bg_light": "#F0FDFA",
        "signal": "#1D4ED8",
    },
    "messaging": {
        "name": "Event Pulse",
        "accent": "#9333EA",
        "accent_soft": "#F3E8FF",
        "background": "#FAF5FF",
        "bg_light": "#FAF5FF",
        "signal": "#EA580C",
    },
    "microservices": {
        "name": "Boundary Mesh",
        "accent": "#0891B2",
        "accent_soft": "#CFFAFE",
        "background": "#F0FDFF",
        "bg_light": "#F0FDFF",
        "signal": "#DC2626",
    },
    "quiz": {
        "name": "Knowledge Pulse",
        "accent": "#7C3AED",
        "accent_soft": "#EDE9FE",
        "background": "#F5F3FF",
        "bg_light": "#F5F3FF",
        "signal": "#0F766E",
    },
}

# Default tech preset accent for detection
DEFAULT_ACCENT_RGB = ImageColor.getrgb("#2563EB")  # Blue
DEFAULT_ACCENT_SOFT_RGB = ImageColor.getrgb("#DBEAFE")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Hue-shift PNGs to folder Visual DNA.")
    parser.add_argument(
        "--folder",
        required=True,
        help="Palette key to use, or 'all' for every configured palette",
    )
    parser.add_argument(
        "--scope",
        default=None,
        help="Relative subtree under assets/go to process recursively. Defaults to the folder name.",
    )
    parser.add_argument("--apply", action="store_true", help="Apply changes (overwrite PNGs)")
    parser.add_argument("--preview", action="store_true", help="Preview mode (create _preview files)")
    parser.add_argument("--base", type=str, default=None,
                        help="Base path to assets/go/ directory")
    return parser.parse_args()


def default_assets_go_base() -> Path:
    return Path(__file__).resolve().parents[4] / "assets" / "go"


def color_distance(c1: tuple[int, int, int], c2: tuple[int, int, int]) -> float:
    """Euclidean distance in RGB space."""
    return ((c1[0]-c2[0])**2 + (c1[1]-c2[1])**2 + (c1[2]-c2[2])**2) ** 0.5


def shift_pixel(r: int, g: int, b: int, 
                source_accent: tuple[int, int, int],
                target_accent: tuple[int, int, int],
                threshold: float = 120.0) -> tuple[int, int, int]:
    """Shift pixel color if it's close to the source accent."""
    dist = color_distance((r, g, b), source_accent)
    if dist < threshold:
        # Blend factor: 1.0 at center, 0.0 at threshold
        factor = 1.0 - (dist / threshold)
        factor = factor ** 0.8  # Soften the blend curve
        nr = int(r + (target_accent[0] - source_accent[0]) * factor)
        ng = int(g + (target_accent[1] - source_accent[1]) * factor)
        nb = int(b + (target_accent[2] - source_accent[2]) * factor)
        return (max(0, min(255, nr)), max(0, min(255, ng)), max(0, min(255, nb)))
    return (r, g, b)


def hue_shift_image(img: Image.Image, palette: dict) -> Image.Image:
    """Apply hue shift from default tech blue to folder-specific accent."""
    target_accent = ImageColor.getrgb(palette["accent"])
    target_soft = ImageColor.getrgb(palette["accent_soft"])
    
    # Work on RGB
    if img.mode != "RGB":
        img = img.convert("RGB")
    
    pixels = img.load()
    width, height = img.size
    
    for y in range(height):
        for x in range(width):
            r, g, b = pixels[x, y]
            
            # Shift main accent (blue -> target)
            r, g, b = shift_pixel(r, g, b, DEFAULT_ACCENT_RGB, target_accent, threshold=100)
            
            # Shift soft accent (light blue -> target soft)
            r, g, b = shift_pixel(r, g, b, DEFAULT_ACCENT_SOFT_RGB, target_soft, threshold=80)
            
            pixels[x, y] = (r, g, b)
    
    return img


def collect_png_files(base_path: Path, scope_name: str) -> list[Path]:
    scope = base_path / scope_name
    if not scope.exists():
        return []
    return sorted(
        path
        for path in scope.rglob("*.png")
        if path.parent.name == "images" and not path.stem.endswith("_preview")
    )


def process_folder(folder_name: str, scope_name: str, base_path: Path, apply: bool, preview: bool) -> None:
    """Process all PNGs in a folder."""
    if folder_name not in FOLDER_PALETTES:
        print(f"Unknown folder: {folder_name}")
        return
    
    palette = FOLDER_PALETTES[folder_name]
    scope_dir = base_path / scope_name

    if not scope_dir.exists():
        print(f"Folder not found: {scope_dir}")
        return

    png_files = collect_png_files(base_path, scope_name)
    print(f"\n{'='*60}")
    print(f"  Palette: {folder_name}")
    print(f"  Scope: {scope_name}")
    print(f"  Palette: {palette['name']}")
    print(f"  Accent: {palette['accent']}")
    print(f"  Files: {len(png_files)}")
    print(f"  Mode: {'APPLY' if apply else 'PREVIEW' if preview else 'DRY RUN'}")
    print(f"{'='*60}\n")
    
    for png in png_files:
        print(f"  Processing: {png.name}...", end=" ")
        try:
            img = Image.open(png)
            shifted = hue_shift_image(img.copy(), palette)
            
            if apply:
                shifted.save(png, format="PNG")
                print("✅ applied")
            elif preview:
                preview_path = png.with_stem(f"{png.stem}_preview")
                shifted.save(preview_path, format="PNG")
                print(f"✅ preview → {preview_path.name}")
            else:
                print("✅ (dry run)")
        except Exception as e:
            print(f"❌ error: {e}")
    
    print(f"\n  Done: {len(png_files)} files processed.\n")


def main() -> None:
    args = parse_args()
    
    if args.base:
        base_path = Path(args.base)
    else:
        base_path = default_assets_go_base()
    
    if not base_path.exists():
        print(f"Base path not found: {base_path}")
        sys.exit(1)
    
    if args.folder == "all":
        if args.scope:
            print("--scope cannot be combined with --folder all")
            sys.exit(1)
        folders = sorted(FOLDER_PALETTES)
    else:
        if args.folder not in FOLDER_PALETTES:
            print(f"Unknown folder palette: {args.folder}")
            sys.exit(1)
        folders = [args.folder]
    
    for folder in folders:
        scope_name = args.scope or folder
        process_folder(folder, scope_name, base_path, args.apply, args.preview)


if __name__ == "__main__":
    main()
