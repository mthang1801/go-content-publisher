#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

from PIL import Image, ImageColor, ImageDraw, ImageFont


CONTENT_PRESETS: dict[str, dict[str, Any]] = {
    "tech": {
        "background": "#F3F7FC",
        "text": "#0F172A",
        "muted": "#475569",
        "accent": "#2563EB",
        "accent_soft": "#DBEAFE",
        "panel": "#FFFFFF",
        "border": "#CBD5E1",
        "shadow": (15, 23, 42, 28),
        "title_regular": [
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
        ],
        "title_bold": [
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
        ],
        "body_regular": [
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
        ],
        "body_bold": [
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
        ],
    },
    "glossary": {
        "background": "#FAF8F4",
        "text": "#2F241F",
        "muted": "#6B5B4E",
        "accent": "#A16207",
        "accent_soft": "#FEF3C7",
        "panel": "#FFFDF9",
        "border": "#E7DCC8",
        "shadow": (64, 48, 32, 24),
        "title_regular": [
            "/usr/share/fonts/google-crosextra-caladea-fonts/Caladea-Regular.ttf",
            "/usr/share/fonts/google-carlito-fonts/Carlito-Regular.ttf",
        ],
        "title_bold": [
            "/usr/share/fonts/google-crosextra-caladea-fonts/Caladea-Bold.ttf",
            "/usr/share/fonts/google-carlito-fonts/Carlito-Bold.ttf",
        ],
        "body_regular": [
            "/usr/share/fonts/google-carlito-fonts/Carlito-Regular.ttf",
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
        ],
        "body_bold": [
            "/usr/share/fonts/google-carlito-fonts/Carlito-Bold.ttf",
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
        ],
    },
    "product": {
        "background": "#FFF8F1",
        "text": "#2C1F17",
        "muted": "#6B4F3F",
        "accent": "#EA580C",
        "accent_soft": "#FFEDD5",
        "panel": "#FFFFFF",
        "border": "#F4C9A3",
        "shadow": (86, 43, 12, 24),
        "title_regular": [
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
        ],
        "title_bold": [
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
        ],
        "body_regular": [
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
        ],
        "body_bold": [
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
        ],
    },
    "game": {
        "background": "#130F22",
        "text": "#F7F4FF",
        "muted": "#D6CDEE",
        "accent": "#F97316",
        "accent_soft": "#3A274A",
        "panel": "#201731",
        "border": "#4B3C66",
        "shadow": (0, 0, 0, 64),
        "title_regular": [
            "/usr/share/fonts/aajohan-comfortaa-fonts/Comfortaa-Regular.otf",
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
        ],
        "title_bold": [
            "/usr/share/fonts/aajohan-comfortaa-fonts/Comfortaa-Bold.otf",
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
        ],
        "body_regular": [
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
        ],
        "body_bold": [
            "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
        ],
    },
    "career": {
        "background": "#F7F4F0",
        "text": "#31231C",
        "muted": "#6E5B50",
        "accent": "#7C3AED",
        "accent_soft": "#EDE9FE",
        "panel": "#FFFEFC",
        "border": "#DDCFC6",
        "shadow": (74, 59, 44, 22),
        "title_regular": [
            "/usr/share/fonts/google-crosextra-caladea-fonts/Caladea-Regular.ttf",
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
        ],
        "title_bold": [
            "/usr/share/fonts/google-crosextra-caladea-fonts/Caladea-Bold.ttf",
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
        ],
        "body_regular": [
            "/usr/share/fonts/google-carlito-fonts/Carlito-Regular.ttf",
            "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
        ],
        "body_bold": [
            "/usr/share/fonts/google-carlito-fonts/Carlito-Bold.ttf",
            "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
        ],
    },
}

MONO_FONT_PATHS = [
    "/usr/share/fonts/google-noto/NotoSansMono-Regular.ttf",
    "/usr/share/fonts/liberation-mono/LiberationMono-Regular.ttf",
    "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
]

VALID_VISUAL_KINDS = {
    "router_map",
    "decision_map",
    "compare_card",
    "taxonomy_card",
    "workflow_map",
    "mental_model_card",
    "api_family_map",
    "boundary_map",
    "state_trace",
    "sequence_flow",
}

VISUAL_KIND_LABELS = {
    "router_map": "ROUTER",
    "decision_map": "DECISION",
    "compare_card": "COMPARE",
    "taxonomy_card": "TAXONOMY",
    "workflow_map": "WORKFLOW",
    "mental_model_card": "MENTAL MODEL",
    "api_family_map": "API MAP",
    "boundary_map": "BOUNDARY MAP",
    "state_trace": "STATE TRACE",
    "sequence_flow": "SEQUENCE FLOW",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Render repo-native documentation visuals to PNG.")
    parser.add_argument("--spec", required=True, help="Path to the input JSON spec.")
    return parser.parse_args()


def load_spec(path: Path) -> dict[str, Any]:
    data = json.loads(path.read_text(encoding="utf-8"))
    required = ["content_type", "visual_kind", "title", "subtitle", "panels", "output_path"]
    missing = [key for key in required if key not in data]
    if missing:
        raise ValueError(f"Missing required keys: {', '.join(missing)}")
    if data["content_type"] not in CONTENT_PRESETS:
        raise ValueError(f"Unsupported content_type: {data['content_type']}")
    if data["visual_kind"] not in VALID_VISUAL_KINDS:
        raise ValueError(f"Unsupported visual_kind: {data['visual_kind']}")
    if not isinstance(data["panels"], list) or not data["panels"]:
        raise ValueError("panels must be a non-empty list")
    palette_overrides = data.get("palette_overrides")
    if palette_overrides is not None:
        if not isinstance(palette_overrides, dict):
            raise ValueError("palette_overrides must be an object")
        allowed_keys = {"background", "text", "muted", "accent", "accent_soft", "panel", "border"}
        invalid_keys = [key for key in palette_overrides if key not in allowed_keys]
        if invalid_keys:
            raise ValueError(f"Unsupported palette_overrides keys: {', '.join(invalid_keys)}")
        for key, value in palette_overrides.items():
            if not isinstance(value, str):
                raise ValueError(f"palette_overrides.{key} must be a string")
            ImageColor.getrgb(value)
    return data


def hex_rgba(value: str, alpha: int = 255) -> tuple[int, int, int, int]:
    r, g, b = ImageColor.getrgb(value)
    return (r, g, b, alpha)


def choose_font(paths: list[str], size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    for raw_path in paths:
        path = Path(raw_path)
        if path.exists():
            return ImageFont.truetype(str(path), size=size)
    return ImageFont.load_default()


def build_fonts(preset: dict[str, Any]) -> dict[str, ImageFont.ImageFont]:
    return {
        "eyebrow": choose_font(preset["body_bold"], 24),
        "title": choose_font(preset["title_bold"], 52),
        "subtitle": choose_font(preset["body_regular"], 28),
        "badge": choose_font(preset["body_bold"], 18),
        "panel_title": choose_font(preset["body_bold"], 24),
        "body": choose_font(preset["body_regular"], 20),
        "body_small": choose_font(preset["body_regular"], 18),
        "body_bold": choose_font(preset["body_bold"], 20),
        "footer": choose_font(preset["body_regular"], 18),
        "step": choose_font(preset["body_bold"], 26),
        "mono": choose_font(MONO_FONT_PATHS, 18),
        "mono_small": choose_font(MONO_FONT_PATHS, 16),
    }


def make_measure_draw() -> ImageDraw.ImageDraw:
    image = Image.new("RGBA", (8, 8), (255, 255, 255, 0))
    return ImageDraw.Draw(image, "RGBA")


def wrap_text(draw: ImageDraw.ImageDraw, text: str, font: ImageFont.ImageFont, max_width: int) -> list[str]:
    words = text.split()
    if not words:
        return []
    lines: list[str] = []
    current = words[0]
    for word in words[1:]:
        trial = f"{current} {word}"
        bbox = draw.textbbox((0, 0), trial, font=font)
        if bbox[2] - bbox[0] <= max_width:
            current = trial
        else:
            lines.append(current)
            current = word
    lines.append(current)
    return lines


def measure_lines(draw: ImageDraw.ImageDraw, lines: list[str], font: ImageFont.ImageFont, line_gap: int = 8) -> int:
    if not lines:
        return 0
    y = 0
    for line in lines:
        text = line if line.strip() else "Ag"
        bbox = draw.textbbox((0, y), text, font=font)
        y = bbox[3] + line_gap
    return y - line_gap



def measure_text_block(
    draw: ImageDraw.ImageDraw,
    text: str,
    font: ImageFont.ImageFont,
    max_width: int,
    line_gap: int,
) -> tuple[list[str], int]:
    lines = wrap_text(draw, text, font, max_width)
    return lines, measure_lines(draw, lines, font, line_gap)


def draw_lines(
    draw: ImageDraw.ImageDraw,
    origin: tuple[int, int],
    lines: list[str],
    font: ImageFont.ImageFont,
    fill: str | tuple[int, int, int, int],
    line_gap: int = 8,
) -> int:
    x, y = origin
    if not lines:
        return 0
    for line in lines:
        draw.text((x, y), line, font=font, fill=fill)
        bbox = draw.textbbox((x, y), line, font=font)
        y = bbox[3] + line_gap
    return y - origin[1] - line_gap


def measure_bullet_block(
    draw: ImageDraw.ImageDraw,
    width: int,
    items: list[str],
    font: ImageFont.ImageFont,
) -> int:
    cursor = 0
    rendered = 0
    for item in items:
        lines = wrap_text(draw, item, font, width - 28)
        if not lines:
            continue
        cursor += measure_lines(draw, lines, font, line_gap=6) + 14
        rendered += 1
    if rendered == 0:
        return 0
    return cursor - 14


def draw_bullet_block(
    draw: ImageDraw.ImageDraw,
    x: int,
    y: int,
    width: int,
    items: list[str],
    font: ImageFont.ImageFont,
    fill: str | tuple[int, int, int, int],
    bullet_fill: str | tuple[int, int, int, int],
) -> int:
    cursor = y
    for item in items:
        lines = wrap_text(draw, item, font, width - 28)
        if not lines:
            continue
        draw.ellipse((x, cursor + 8, x + 10, cursor + 18), fill=bullet_fill)
        block_height = draw_lines(draw, (x + 20, cursor), lines, font, fill, line_gap=6)
        cursor += block_height + 14
    return cursor - y


def measure_card_height(
    draw: ImageDraw.ImageDraw,
    fonts: dict[str, ImageFont.ImageFont],
    card_width: int,
    heading: str,
    body: str = "",
    bullets: list[str] | None = None,
    tag: str = "",
) -> int:
    content_x = 52
    max_width = card_width - content_x - 26
    reserved_heading_width = 0

    if tag:
        tag_box = draw.textbbox((0, 0), tag, font=fonts["badge"])
        tag_width = tag_box[2] - tag_box[0] + 30
        reserved_heading_width = min(tag_width + 18, max_width // 2)

    heading_lines = wrap_text(draw, heading, fonts["panel_title"], max_width - reserved_heading_width)
    total_height = 28
    total_height += measure_lines(draw, heading_lines, fonts["panel_title"], line_gap=6)
    total_height += 12

    if body:
        body_lines = wrap_text(draw, body, fonts["body"], max_width)
        total_height += measure_lines(draw, body_lines, fonts["body"], line_gap=7)
        total_height += 12

    if bullets:
        total_height += measure_bullet_block(draw, max_width, bullets, fonts["body_small"])

    total_height += 36
    return max(170, total_height)


def draw_card(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    heading: str,
    body: str = "",
    bullets: list[str] | None = None,
    tag: str = "",
) -> None:
    x0, y0, x1, y1 = box
    shadow_box = (x0 + 8, y0 + 10, x1 + 8, y1 + 10)
    draw.rounded_rectangle(shadow_box, radius=28, fill=preset["shadow"])
    draw.rounded_rectangle(box, radius=28, fill=preset["panel"], outline=preset["border"], width=2)
    draw.rounded_rectangle((x0 + 18, y0 + 18, x0 + 30, y1 - 18), radius=8, fill=preset["accent"])
    content_x = x0 + 52
    cursor_y = y0 + 28
    max_width = x1 - content_x - 26
    reserved_heading_width = 0

    if tag:
        tag_box = draw.textbbox((0, 0), tag, font=fonts["badge"])
        tag_width = tag_box[2] - tag_box[0] + 30
        reserved_heading_width = min(tag_width + 18, max_width // 2)
        draw.rounded_rectangle(
            (x1 - tag_width - 22, y0 + 20, x1 - 22, y0 + 56),
            radius=18,
            fill=preset["accent_soft"],
        )
        draw.text((x1 - tag_width - 7, y0 + 25), tag, font=fonts["badge"], fill=preset["accent"])

    heading_lines = wrap_text(draw, heading, fonts["panel_title"], max_width - reserved_heading_width)
    cursor_y += draw_lines(draw, (content_x, cursor_y), heading_lines, fonts["panel_title"], preset["text"], line_gap=6)
    cursor_y += 12

    if body:
        body_lines = wrap_text(draw, body, fonts["body"], max_width)
        cursor_y += draw_lines(draw, (content_x, cursor_y), body_lines, fonts["body"], preset["muted"], line_gap=7)
        cursor_y += 12

    if bullets:
        draw_bullet_block(draw, content_x, cursor_y, max_width, bullets, fonts["body_small"], preset["text"], preset["accent"])


def draw_lane_label(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    text: str,
    preset: dict[str, Any],
    font: ImageFont.ImageFont,
) -> None:
    if not text:
        return
    text_box = draw.textbbox((0, 0), text, font=font)
    text_width = text_box[2] - text_box[0]
    pill_width = text_width + 28
    x0, y0, _, _ = box
    draw.rounded_rectangle(
        (x0, y0, x0 + pill_width, y0 + 36),
        radius=18,
        fill=preset["accent_soft"],
        outline=hex_rgba(preset["accent"], 72),
        width=2,
    )
    draw.text((x0 + 14, y0 + 8), text, font=font, fill=preset["accent"])


def measure_state_block(
    draw: ImageDraw.ImageDraw,
    width: int,
    lines: list[str],
    font: ImageFont.ImageFont,
) -> int:
    if not lines:
        return 0
    text_height = measure_lines(draw, lines, font, line_gap=6)
    return text_height + 34


def draw_state_block(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    lines: list[str],
    preset: dict[str, Any],
    font: ImageFont.ImageFont,
) -> None:
    if not lines:
        return
    x0, y0, x1, y1 = box
    draw.rounded_rectangle(
        box,
        radius=18,
        fill=hex_rgba(preset["accent"], 16),
        outline=hex_rgba(preset["accent"], 60),
        width=2,
    )
    draw_lines(draw, (x0 + 18, y0 + 16), lines, font, preset["text"], line_gap=6)


def connector_points(box: tuple[int, int, int, int], side: str) -> tuple[int, int]:
    x0, y0, x1, y1 = box
    if side == "left":
        return (x1, (y0 + y1) // 2)
    return (x0, (y0 + y1) // 2)


def draw_connector(
    draw: ImageDraw.ImageDraw,
    start: tuple[int, int],
    end: tuple[int, int],
    color: str,
    direction: str = "right",
) -> None:
    sx, sy = start
    ex, ey = end
    mid_x = int((sx + ex) / 2)
    draw.line((sx, sy, mid_x, sy), fill=color, width=5)
    draw.line((mid_x, sy, mid_x, ey), fill=color, width=5)
    draw.line((mid_x, ey, ex, ey), fill=color, width=5)
    if direction == "right":
        arrow = ((ex - 16, ey - 10), (ex - 16, ey + 10), (ex + 2, ey))
    else:
        arrow = ((ex + 16, ey - 10), (ex + 16, ey + 10), (ex - 2, ey))
    draw.polygon(arrow, fill=color)


def measure_header_bottom(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    margin = 72
    header_top = 64
    if spec.get("eyebrow"):
        header_top += 42
    title_lines = wrap_text(draw, spec["title"], fonts["title"], width - margin * 2 - 220)
    title_height = measure_lines(draw, title_lines, fonts["title"], line_gap=4)
    subtitle_y = header_top + title_height + 16
    subtitle_lines = wrap_text(draw, spec["subtitle"], fonts["subtitle"], width - margin * 2 - 260)
    subtitle_height = measure_lines(draw, subtitle_lines, fonts["subtitle"], line_gap=6)
    return subtitle_y + subtitle_height + 28


def background_decor(draw: ImageDraw.ImageDraw, width: int, height: int, preset: dict[str, Any]) -> None:
    draw.rectangle((0, 0, width, height), fill=preset["background"])
    draw.ellipse((-120, -80, 360, 300), fill=hex_rgba(preset["accent"], 26))
    draw.ellipse((width - 320, -60, width + 80, 240), fill=hex_rgba(preset["accent"], 18))
    draw.ellipse((width - 260, height - 240, width + 120, height + 80), fill=hex_rgba(preset["accent"], 12))


def draw_header(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> tuple[int, int, int, int]:
    margin = 72
    header_top = 64
    if spec.get("eyebrow"):
        draw.text((margin, header_top), spec["eyebrow"].upper(), font=fonts["eyebrow"], fill=preset["accent"])
        header_top += 42
    title_lines = wrap_text(draw, spec["title"], fonts["title"], width - margin * 2 - 220)
    title_height = draw_lines(draw, (margin, header_top), title_lines, fonts["title"], preset["text"], line_gap=4)
    subtitle_y = header_top + title_height + 16
    subtitle_lines = wrap_text(draw, spec["subtitle"], fonts["subtitle"], width - margin * 2 - 260)
    subtitle_height = draw_lines(draw, (margin, subtitle_y), subtitle_lines, fonts["subtitle"], preset["muted"], line_gap=6)

    visual_label = VISUAL_KIND_LABELS.get(spec["visual_kind"], spec["visual_kind"].replace("_", " ").upper())
    badge = f"{spec['content_type']} / {visual_label}".upper()
    badge_box = draw.textbbox((0, 0), badge, font=fonts["badge"])
    badge_width = badge_box[2] - badge_box[0] + 32
    badge_height = badge_box[3] - badge_box[1] + 20
    badge_x = width - margin - badge_width
    badge_y = 72
    draw.rounded_rectangle((badge_x, badge_y, badge_x + badge_width, badge_y + badge_height), radius=20, fill=preset["panel"], outline=preset["border"], width=2)
    draw.text((badge_x + 16, badge_y + 9), badge.upper(), font=fonts["badge"], fill=preset["accent"])

    return (margin, subtitle_y + subtitle_height + 28, width - margin, 0)


def draw_grid_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
    columns: int,
) -> None:
    x0, y0, x1, _ = area
    panels = spec["panels"]
    gutter = 22
    usable_width = x1 - x0
    card_width = int((usable_width - gutter * (columns - 1)) / columns)
    rows = (len(panels) + columns - 1) // columns
    row_heights: list[int] = []

    for row in range(rows):
        start = row * columns
        row_panels = panels[start : start + columns]
        row_height = max(
            measure_card_height(
                draw,
                fonts,
                card_width,
                heading=panel.get("heading", ""),
                body=panel.get("body", ""),
                bullets=panel.get("bullets", []),
                tag=panel.get("tag", ""),
            )
            for panel in row_panels
        )
        row_heights.append(row_height)

    current_y = y0
    for row, row_height in enumerate(row_heights):
        row_panels = panels[row * columns : (row + 1) * columns]
        for col, panel in enumerate(row_panels):
            card_x0 = x0 + col * (card_width + gutter)
            box = (card_x0, current_y, card_x0 + card_width, current_y + row_height)
            draw_card(
                draw,
                box,
                preset,
                fonts,
                heading=panel.get("heading", ""),
                body=panel.get("body", ""),
                bullets=panel.get("bullets", []),
                tag=panel.get("tag", ""),
            )
        current_y += row_height + gutter


def draw_compare_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, _ = area
    panels = spec["panels"]
    if len(panels) < 2:
        draw_grid_layout(draw, spec, preset, fonts, area, columns=1)
        return
    gutter = 24

    col_width = int((x1 - x0 - gutter) / 2)
    main_height = max(
        measure_card_height(draw, fonts, col_width, panels[0].get("heading", ""), panels[0].get("body", ""), panels[0].get("bullets", []), panels[0].get("tag", "A")),
        measure_card_height(draw, fonts, col_width, panels[1].get("heading", ""), panels[1].get("body", ""), panels[1].get("bullets", []), panels[1].get("tag", "B")),
    )
    left_box = (x0, y0, x0 + col_width, y0 + main_height)
    right_box = (x0 + col_width + gutter, y0, x1, y0 + main_height)

    draw_card(draw, left_box, preset, fonts, panels[0].get("heading", ""), panels[0].get("body", ""), panels[0].get("bullets", []), panels[0].get("tag", "A"))
    draw_card(draw, right_box, preset, fonts, panels[1].get("heading", ""), panels[1].get("body", ""), panels[1].get("bullets", []), panels[1].get("tag", "B"))

    if len(panels) > 2:
        rest = panels[2:]
        foot_area = (x0, y0 + main_height + gutter, x1, y0 + main_height + gutter)
        draw_grid_layout(draw, {**spec, "panels": rest}, preset, fonts, foot_area, columns=min(len(rest), 3))


def draw_workflow_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, _ = area
    panels = spec["panels"]
    count = len(panels)
    gutter = 18
    card_width = int((x1 - x0 - gutter * (count - 1)) / count)
    card_height = max(
        measure_card_height(
            draw,
            fonts,
            card_width,
            panel.get("heading", f"Step {index + 1}"),
            panel.get("body", ""),
            panel.get("bullets", []),
            panel.get("tag", f"{index + 1:02d}"),
        )
        for index, panel in enumerate(panels)
    )

    for index, panel in enumerate(panels):
        cx0 = x0 + index * (card_width + gutter)
        cx1 = cx0 + card_width
        box = (cx0, y0, cx1, y0 + card_height)
        draw_card(draw, box, preset, fonts, panel.get("heading", f"Step {index + 1}"), panel.get("body", ""), panel.get("bullets", []), panel.get("tag", f"{index + 1:02d}"))
        if index < count - 1:
            start_x = cx1 + 8
            mid_y = y0 + card_height // 2
            end_x = cx1 + gutter - 8
            draw.line((start_x, mid_y, end_x, mid_y), fill=preset["accent"], width=6)
            draw.polygon(((end_x - 14, mid_y - 10), (end_x - 14, mid_y + 10), (end_x + 2, mid_y)), fill=preset["accent"])


def draw_mental_model_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, _ = area
    hero_lines = wrap_text(draw, spec["subtitle"], fonts["body_bold"], x1 - x0 - 64)
    hero_height = max(108, measure_lines(draw, hero_lines, fonts["body_bold"], line_gap=6) + 76)
    draw.rounded_rectangle((x0, y0, x1, y0 + hero_height), radius=28, fill=preset["panel"], outline=preset["border"], width=2)
    draw_lines(draw, (x0 + 32, y0 + 38), hero_lines, fonts["body_bold"], preset["accent"], line_gap=6)
    rest_area = (x0, y0 + hero_height + 24, x1, y0 + hero_height + 24)
    draw_grid_layout(draw, spec, preset, fonts, rest_area, columns=2 if len(spec["panels"]) > 1 else 1)


def draw_decision_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, _ = area
    prompt_lines = wrap_text(draw, "Choose based on the condition that matches first.", fonts["body_bold"], x1 - x0 - 360)
    prompt_height = max(110, measure_lines(draw, prompt_lines, fonts["body_bold"], line_gap=6) + 64)
    prompt_box = (x0 + 160, y0, x1 - 160, y0 + prompt_height)
    draw.rounded_rectangle(prompt_box, radius=24, fill=preset["accent_soft"], outline=preset["border"], width=2)
    draw_lines(draw, (prompt_box[0] + 24, prompt_box[1] + 32), prompt_lines, fonts["body_bold"], preset["accent"], line_gap=6)

    panels_area = (x0, y0 + prompt_height + 32, x1, y0 + prompt_height + 32)
    columns = 2 if len(spec["panels"]) >= 4 else max(1, len(spec["panels"]))
    draw_grid_layout(draw, spec, preset, fonts, panels_area, columns=columns)


def split_boundary_panels(spec: dict[str, Any]) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    left_panels = [panel for panel in spec["panels"] if panel.get("side", "left") == "left"]
    right_panels = [panel for panel in spec["panels"] if panel.get("side") == "right"]
    if not left_panels and not right_panels:
        midpoint = (len(spec["panels"]) + 1) // 2
        return spec["panels"][:midpoint], spec["panels"][midpoint:]
    if not left_panels:
        midpoint = max(1, len(right_panels) // 2)
        return right_panels[:midpoint], right_panels[midpoint:]
    if not right_panels:
        midpoint = max(1, len(left_panels) // 2)
        return left_panels[:midpoint], left_panels[midpoint:]
    return left_panels, right_panels


def measure_boundary_layout_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    gutter = 26
    lane_width = max(280, int((width - 360 - gutter * 2) / 2))
    left_panels, right_panels = split_boundary_panels(spec)
    lane_heights: list[int] = []
    for panels in (left_panels, right_panels):
        lane_height = 0
        for index, panel in enumerate(panels):
            lane_height += measure_card_height(
                draw,
                fonts,
                lane_width,
                panel.get("heading", ""),
                panel.get("body", ""),
                panel.get("bullets", []),
                panel.get("tag", ""),
            )
            if index < len(panels) - 1:
                lane_height += gutter
        lane_heights.append(lane_height)

    core = spec.get("core", {})
    core_height = measure_card_height(
        draw,
        fonts,
        320,
        core.get("heading", "Application core"),
        core.get("body", ""),
        core.get("bullets", []),
        core.get("tag", "CORE"),
    )
    return max(core_height, *lane_heights)


def draw_boundary_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, _ = area
    left_panels, right_panels = split_boundary_panels(spec)
    gutter = 26
    center_width = 320
    lane_width = max(280, int((x1 - x0 - center_width - gutter * 2) / 2))
    left_x = x0
    core_x = left_x + lane_width + gutter
    right_x = core_x + center_width + gutter

    lane_label_y = y0
    draw_lane_label(draw, (left_x, lane_label_y, 0, 0), spec.get("left_label", "Driver side").upper(), preset, fonts["badge"])
    draw_lane_label(draw, (right_x, lane_label_y, 0, 0), spec.get("right_label", "Driven side").upper(), preset, fonts["badge"])

    content_y = y0 + 52
    left_boxes: list[tuple[int, int, int, int]] = []
    cursor = content_y
    for panel in left_panels:
        card_height = measure_card_height(draw, fonts, lane_width, panel.get("heading", ""), panel.get("body", ""), panel.get("bullets", []), panel.get("tag", ""))
        box = (left_x, cursor, left_x + lane_width, cursor + card_height)
        draw_card(draw, box, preset, fonts, panel.get("heading", ""), panel.get("body", ""), panel.get("bullets", []), panel.get("tag", ""))
        left_boxes.append(box)
        cursor += card_height + gutter

    right_boxes: list[tuple[int, int, int, int]] = []
    cursor = content_y
    for panel in right_panels:
        card_height = measure_card_height(draw, fonts, lane_width, panel.get("heading", ""), panel.get("body", ""), panel.get("bullets", []), panel.get("tag", ""))
        box = (right_x, cursor, right_x + lane_width, cursor + card_height)
        draw_card(draw, box, preset, fonts, panel.get("heading", ""), panel.get("body", ""), panel.get("bullets", []), panel.get("tag", ""))
        right_boxes.append(box)
        cursor += card_height + gutter

    core = spec.get("core", {})
    core_height = measure_card_height(
        draw,
        fonts,
        center_width,
        core.get("heading", "Application core"),
        core.get("body", ""),
        core.get("bullets", []),
        core.get("tag", "CORE"),
    )
    max_lane_bottom = max([content_y + core_height, *[box[3] for box in left_boxes], *[box[3] for box in right_boxes]])
    core_y = content_y + max(0, (max_lane_bottom - content_y - core_height) // 2)
    core_box = (core_x, core_y, core_x + center_width, core_y + core_height)
    draw_card(
        draw,
        core_box,
        preset,
        fonts,
        core.get("heading", "Application core"),
        core.get("body", ""),
        core.get("bullets", []),
        core.get("tag", "CORE"),
    )

    core_left = connector_points(core_box, "right")
    core_right = connector_points(core_box, "left")
    for box in left_boxes:
        draw_connector(draw, connector_points(box, "left"), core_left, preset["accent"], direction="right")
    for box in right_boxes:
        draw_connector(draw, connector_points(box, "right"), core_right, preset["accent"], direction="left")


def measure_state_trace_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    step_width = width - 88
    total = 0
    for index, panel in enumerate(spec["panels"]):
        body_height = 0
        if panel.get("body"):
            _, body_height = measure_text_block(draw, panel["body"], fonts["body"], step_width - 84, line_gap=7)
        bullet_height = measure_bullet_block(draw, step_width - 84, panel.get("bullets", []), fonts["body_small"])
        state_lines = panel.get("state_lines", [])
        wrapped_state_lines: list[str] = []
        for line in state_lines:
            wrapped_state_lines.extend(wrap_text(draw, line, fonts["mono_small"], step_width - 120))
        state_height = measure_state_block(draw, step_width - 84, wrapped_state_lines, fonts["mono_small"])
        total += max(164, 88 + body_height + (12 if body_height and bullet_height else 0) + bullet_height + (18 if state_height else 0) + state_height)
        if index < len(spec["panels"]) - 1:
            total += 18
    return total


def draw_state_trace_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, _ = area
    rail_x = x0 + 28
    step_x = x0 + 72
    step_width = x1 - step_x
    cursor = y0

    for index, panel in enumerate(spec["panels"]):
        heading = panel.get("heading", f"Step {index + 1}")
        body = panel.get("body", "")
        bullets = panel.get("bullets", [])
        state_lines = panel.get("state_lines", [])
        wrapped_state_lines: list[str] = []
        for line in state_lines:
            wrapped_state_lines.extend(wrap_text(draw, line, fonts["mono_small"], step_width - 120))

        body_lines = wrap_text(draw, body, fonts["body"], step_width - 84) if body else []
        body_height = measure_lines(draw, body_lines, fonts["body"], line_gap=7) if body_lines else 0
        bullet_height = measure_bullet_block(draw, step_width - 84, bullets, fonts["body_small"])
        state_height = measure_state_block(draw, step_width - 84, wrapped_state_lines, fonts["mono_small"])
        card_height = max(164, 88 + body_height + (12 if body_height and bullet_height else 0) + bullet_height + (18 if state_height else 0) + state_height)

        draw.rounded_rectangle(
            (step_x, cursor, step_x + step_width, cursor + card_height),
            radius=28,
            fill=preset["panel"],
            outline=preset["border"],
            width=2,
        )
        draw.ellipse((rail_x - 12, cursor + 26, rail_x + 28, cursor + 66), fill=preset["accent"])
        draw.text((rail_x - 1, cursor + 35), f"{index + 1}", font=fonts["badge"], fill=preset["panel"])
        if index < len(spec["panels"]) - 1:
            draw.line((rail_x + 8, cursor + 66, rail_x + 8, cursor + card_height + 18), fill=hex_rgba(preset["accent"], 120), width=4)

        inner_x = step_x + 28
        inner_y = cursor + 24
        tag = panel.get("tag", "")
        if tag:
            draw_lane_label(draw, (step_x + step_width - 200, cursor + 22, 0, 0), tag.upper(), preset, fonts["badge"])
        heading_lines = wrap_text(draw, heading, fonts["panel_title"], step_width - 84 - 120)
        inner_y += draw_lines(draw, (inner_x, inner_y), heading_lines, fonts["panel_title"], preset["text"], line_gap=6)
        if body_lines:
            inner_y += 12
            inner_y += draw_lines(draw, (inner_x, inner_y), body_lines, fonts["body"], preset["muted"], line_gap=7)
        if bullets:
            inner_y += 12
            inner_y += draw_bullet_block(draw, inner_x, inner_y, step_width - 84, bullets, fonts["body_small"], preset["text"], preset["accent"])
        if wrapped_state_lines:
            inner_y += 18
            draw_state_block(draw, (inner_x, inner_y, step_x + step_width - 28, inner_y + state_height), wrapped_state_lines, preset, fonts["mono_small"])

        cursor += card_height + 18


def resolve_actor_index(actor_names: list[str], actor: Any) -> int:
    if isinstance(actor, int):
        if 0 <= actor < len(actor_names):
            return actor
        raise ValueError(f"Actor index out of range: {actor}")
    if isinstance(actor, str):
        if actor in actor_names:
            return actor_names.index(actor)
        raise ValueError(f"Unknown actor name: {actor}")
    raise ValueError(f"Unsupported actor reference: {actor!r}")


def measure_sequence_step_height(
    draw: ImageDraw.ImageDraw,
    panel: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    note_width: int,
) -> int:
    heading_lines = wrap_text(draw, panel.get("heading", ""), fonts["body_bold"], note_width)
    total = max(28, measure_lines(draw, heading_lines, fonts["body_bold"], line_gap=5))
    body = panel.get("body", "")
    if body:
        _, body_height = measure_text_block(draw, body, fonts["body_small"], note_width, line_gap=6)
        total += 10 + body_height
    bullets = panel.get("bullets", [])
    if bullets:
        total += 12 + measure_bullet_block(draw, note_width, bullets, fonts["body_small"])
    return max(110, total + 48)


def measure_sequence_flow_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    note_width = min(360, max(240, width // 4))
    total = 96
    for panel in spec["panels"]:
        total += measure_sequence_step_height(draw, panel, fonts, note_width) + 12
    return total


def draw_arrow(
    draw: ImageDraw.ImageDraw,
    x0: int,
    x1: int,
    y: int,
    color: str,
) -> None:
    draw.line((x0, y, x1, y), fill=color, width=5)
    if x1 >= x0:
        arrow = ((x1 - 16, y - 10), (x1 - 16, y + 10), (x1 + 2, y))
    else:
        arrow = ((x1 + 16, y - 10), (x1 + 16, y + 10), (x1 - 2, y))
    draw.polygon(arrow, fill=color)


def draw_sequence_flow_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    area: tuple[int, int, int, int],
) -> None:
    x0, y0, x1, y1 = area
    actors = spec.get("actors")
    if not isinstance(actors, list) or len(actors) < 2:
        raise ValueError("sequence_flow requires an 'actors' list with at least 2 entries")

    lane_left = x0 + 36
    lane_right = x1 - 36
    if len(actors) == 1:
        actor_x = [int((lane_left + lane_right) / 2)]
    else:
        gap = (lane_right - lane_left) / (len(actors) - 1)
        actor_x = [int(lane_left + gap * idx) for idx in range(len(actors))]

    header_y = y0
    lane_y0 = y0 + 64
    note_width = min(360, max(240, (x1 - x0) // 4))

    for index, actor in enumerate(actors):
        label = str(actor)
        text_box = draw.textbbox((0, 0), label, font=fonts["badge"])
        pill_width = text_box[2] - text_box[0] + 30
        pill_x0 = actor_x[index] - pill_width // 2
        draw.rounded_rectangle(
            (pill_x0, header_y, pill_x0 + pill_width, header_y + 38),
            radius=19,
            fill=preset["panel"],
            outline=preset["border"],
            width=2,
        )
        draw.text((pill_x0 + 15, header_y + 9), label, font=fonts["badge"], fill=preset["accent"])
        draw.line((actor_x[index], lane_y0, actor_x[index], y1 - 12), fill=hex_rgba(preset["border"], 160), width=3)

    cursor = lane_y0 + 10
    for panel in spec["panels"]:
        start_idx = resolve_actor_index(actors, panel.get("from", 0))
        end_idx = resolve_actor_index(actors, panel.get("to", 1))
        start_x = actor_x[start_idx]
        end_x = actor_x[end_idx]
        step_height = measure_sequence_step_height(draw, panel, fonts, note_width)
        line_y = cursor + 28

        heading = panel.get("heading", "")
        heading_box = draw.textbbox((0, 0), heading, font=fonts["body_bold"])
        heading_width = heading_box[2] - heading_box[0] + 24
        heading_x0 = int((start_x + end_x) / 2) - heading_width // 2
        draw.rounded_rectangle(
            (heading_x0, cursor, heading_x0 + heading_width, cursor + 34),
            radius=17,
            fill=preset["accent_soft"],
        )
        draw.text((heading_x0 + 12, cursor + 8), heading, font=fonts["body_bold"], fill=preset["accent"])
        draw_arrow(draw, start_x, end_x, line_y + 20, preset["accent"])

        body = panel.get("body", "")
        bullets = panel.get("bullets", [])
        if body or bullets:
            note_x0 = max(x0 + 12, min(int((start_x + end_x) / 2) - note_width // 2, x1 - note_width - 12))
            note_box = (note_x0, cursor + 48, note_x0 + note_width, cursor + step_height)
            draw.rounded_rectangle(
                note_box,
                radius=20,
                fill=preset["panel"],
                outline=preset["border"],
                width=2,
            )
            inner_x = note_x0 + 18
            inner_y = cursor + 64
            if body:
                body_lines = wrap_text(draw, body, fonts["body_small"], note_width - 36)
                inner_y += draw_lines(draw, (inner_x, inner_y), body_lines, fonts["body_small"], preset["muted"], line_gap=6)
            if bullets:
                if body:
                    inner_y += 10
                draw_bullet_block(draw, inner_x, inner_y, note_width - 36, bullets, fonts["body_small"], preset["text"], preset["accent"])

        cursor += step_height + 12


def measure_grid_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
    columns: int,
    gutter: int = 22,
) -> int:
    panels = spec["panels"]
    card_width = int((width - gutter * (columns - 1)) / columns)
    rows = (len(panels) + columns - 1) // columns
    row_heights = []
    for row in range(rows):
        row_panels = panels[row * columns : (row + 1) * columns]
        row_heights.append(
            max(
                measure_card_height(
                    draw,
                    fonts,
                    card_width,
                    panel.get("heading", ""),
                    panel.get("body", ""),
                    panel.get("bullets", []),
                    panel.get("tag", ""),
                )
                for panel in row_panels
            )
        )
    return sum(row_heights) + gutter * max(0, rows - 1)


def measure_compare_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    panels = spec["panels"]
    if len(panels) < 2:
        return measure_grid_height(draw, spec, fonts, width, columns=1)
    gutter = 24
    col_width = int((width - gutter) / 2)
    main_height = max(
        measure_card_height(draw, fonts, col_width, panels[0].get("heading", ""), panels[0].get("body", ""), panels[0].get("bullets", []), panels[0].get("tag", "A")),
        measure_card_height(draw, fonts, col_width, panels[1].get("heading", ""), panels[1].get("body", ""), panels[1].get("bullets", []), panels[1].get("tag", "B")),
    )
    if len(panels) == 2:
        return main_height
    rest = {**spec, "panels": panels[2:]}
    return main_height + gutter + measure_grid_height(draw, rest, fonts, width, columns=min(len(panels) - 2, 3))


def measure_workflow_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    panels = spec["panels"]
    gutter = 18
    card_width = int((width - gutter * (len(panels) - 1)) / len(panels))
    return max(
        measure_card_height(
            draw,
            fonts,
            card_width,
            panel.get("heading", f"Step {index + 1}"),
            panel.get("body", ""),
            panel.get("bullets", []),
            panel.get("tag", f"{index + 1:02d}"),
        )
        for index, panel in enumerate(panels)
    )


def measure_mental_model_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    hero_lines, hero_text_height = measure_text_block(draw, spec["subtitle"], fonts["body_bold"], width - 64, line_gap=6)
    hero_height = max(108, hero_text_height + 76)
    rest_height = measure_grid_height(draw, spec, fonts, width, columns=2 if len(spec["panels"]) > 1 else 1)
    return hero_height + 24 + rest_height


def measure_decision_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
) -> int:
    _, prompt_text_height = measure_text_block(
        draw,
        "Choose based on the condition that matches first.",
        fonts["body_bold"],
        width - 360,
        line_gap=6,
    )
    prompt_height = max(110, prompt_text_height + 64)
    columns = 2 if len(spec["panels"]) >= 4 else max(1, len(spec["panels"]))
    return prompt_height + 32 + measure_grid_height(draw, spec, fonts, width, columns=columns)


def measure_layout_height(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    body_width: int,
) -> int:
    visual_kind = spec["visual_kind"]
    if visual_kind in {"router_map", "taxonomy_card", "api_family_map"}:
        columns = 3 if len(spec["panels"]) >= 5 else max(1, min(2, len(spec["panels"])))
        return measure_grid_height(draw, spec, fonts, body_width, columns=columns)
    if visual_kind == "boundary_map":
        return measure_boundary_layout_height(draw, spec, fonts, body_width)
    if visual_kind == "compare_card":
        return measure_compare_height(draw, spec, fonts, body_width)
    if visual_kind == "workflow_map":
        return measure_workflow_height(draw, spec, fonts, body_width)
    if visual_kind == "mental_model_card":
        return measure_mental_model_height(draw, spec, fonts, body_width)
    if visual_kind == "decision_map":
        return measure_decision_height(draw, spec, fonts, body_width)
    if visual_kind == "state_trace":
        return measure_state_trace_height(draw, spec, fonts, body_width)
    if visual_kind == "sequence_flow":
        return measure_sequence_flow_height(draw, spec, fonts, body_width)
    return measure_grid_height(draw, spec, fonts, body_width, columns=2)


def resolve_canvas_height(spec: dict[str, Any], preset: dict[str, Any], fonts: dict[str, ImageFont.ImageFont], width: int, requested_height: int) -> int:
    draw = make_measure_draw()
    margin = 72
    body_top = measure_header_bottom(draw, spec, fonts, width)
    footer_space = 68 if spec.get("footer") else 0
    body_height = measure_layout_height(draw, spec, fonts, width - margin * 2)
    resolved_height = max(requested_height, body_top + 18 + body_height + margin + footer_space)
    max_height = int(spec.get("max_height", 2200))
    if resolved_height > max_height:
        raise ValueError(
            f"Resolved height {resolved_height}px exceeds max_height {max_height}px for {spec['output_path']}. "
            "Reduce copy density or split the visual."
        )
    return resolved_height


def draw_layout(
    draw: ImageDraw.ImageDraw,
    spec: dict[str, Any],
    preset: dict[str, Any],
    fonts: dict[str, ImageFont.ImageFont],
    width: int,
    height: int,
) -> None:
    margin = 72
    _, body_top, _, _ = draw_header(draw, spec, preset, fonts, width)
    footer_present = bool(spec.get("footer"))
    footer_space = 68 if footer_present else 0
    body_area = (margin, body_top + 18, width - margin, height - margin - footer_space)

    visual_kind = spec["visual_kind"]
    if visual_kind in {"router_map", "taxonomy_card", "api_family_map"}:
        columns = 3 if len(spec["panels"]) >= 5 else max(1, min(2, len(spec["panels"])))
        draw_grid_layout(draw, spec, preset, fonts, body_area, columns=columns)
    elif visual_kind == "boundary_map":
        draw_boundary_layout(draw, spec, preset, fonts, body_area)
    elif visual_kind == "compare_card":
        draw_compare_layout(draw, spec, preset, fonts, body_area)
    elif visual_kind == "workflow_map":
        draw_workflow_layout(draw, spec, preset, fonts, body_area)
    elif visual_kind == "mental_model_card":
        draw_mental_model_layout(draw, spec, preset, fonts, body_area)
    elif visual_kind == "decision_map":
        draw_decision_layout(draw, spec, preset, fonts, body_area)
    elif visual_kind == "state_trace":
        draw_state_trace_layout(draw, spec, preset, fonts, body_area)
    elif visual_kind == "sequence_flow":
        draw_sequence_flow_layout(draw, spec, preset, fonts, body_area)
    else:
        draw_grid_layout(draw, spec, preset, fonts, body_area, columns=2)

    if spec.get("footer"):
        footer_lines = wrap_text(draw, spec["footer"], fonts["footer"], width - margin * 2)
        draw_lines(draw, (margin, height - margin + 8), footer_lines, fonts["footer"], preset["muted"], line_gap=4)


def render(spec: dict[str, Any]) -> Image.Image:
    width = int(spec.get("width", 1600))
    requested_height = int(spec.get("height", 1000))
    preset = dict(CONTENT_PRESETS[spec["content_type"]])
    preset.update(spec.get("palette_overrides", {}))
    fonts = build_fonts(preset)
    height = resolve_canvas_height(spec, preset, fonts, width, requested_height)
    image = Image.new("RGBA", (width, height), preset["background"])
    draw = ImageDraw.Draw(image, "RGBA")
    background_decor(draw, width, height, preset)
    draw_layout(draw, spec, preset, fonts, width, height)
    return image.convert("RGB")


def main() -> None:
    args = parse_args()
    spec_path = Path(args.spec)
    spec = load_spec(spec_path)
    output_path = Path(spec["output_path"])
    if not output_path.is_absolute():
        output_path = Path.cwd() / output_path
    output_path.parent.mkdir(parents=True, exist_ok=True)
    image = render(spec)
    image.save(output_path, format="PNG")
    print(
        json.dumps(
            {
                "status": "ok",
                "output_path": str(output_path),
                "width": image.width,
                "height": image.height,
                "requested_height": int(spec.get("height", 1000)),
                "content_type": spec["content_type"],
                "visual_kind": spec["visual_kind"],
            },
            ensure_ascii=False,
        )
    )


if __name__ == "__main__":
    main()
