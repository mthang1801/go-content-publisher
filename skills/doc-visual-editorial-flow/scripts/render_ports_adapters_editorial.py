#!/usr/bin/env python3
from __future__ import annotations

import argparse
from pathlib import Path

from PIL import Image, ImageColor, ImageDraw, ImageFont


WIDTH = 1600
HEIGHT = 1120
BG = "#FBF8F2"
TEXT = "#1D2430"
MUTED = "#55606C"
LINE = "#3B4654"
ACCENT = "#D16D43"

PALETTE = {
    "driver_actor": "#F4E9FF",
    "driver_adapter": "#F8C9D1",
    "inbound": "#93C5FD",
    "core": "#F6F0D7",
    "outbound": "#BBF7D0",
    "driven_adapter": "#F7D974",
    "accent_blue": "#4B78B8",
    "accent_orange": "#D16D43",
}

FONT_SANS = [
    "/usr/share/fonts/google-noto/NotoSans-Regular.ttf",
    "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Regular.otf",
]
FONT_SANS_BOLD = [
    "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
    "/usr/share/fonts/abattis-cantarell-fonts/Cantarell-Bold.otf",
]


def rgba(color: str, alpha: int) -> tuple[int, int, int, int]:
    r, g, b = ImageColor.getrgb(color)
    return (r, g, b, alpha)


def choose_font(paths: list[str], size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    for raw_path in paths:
        path = Path(raw_path)
        if path.exists():
            return ImageFont.truetype(str(path), size=size)
    return ImageFont.load_default()


def wrap_text(draw: ImageDraw.ImageDraw, text: str, font: ImageFont.ImageFont, max_width: int) -> list[str]:
    words = text.split()
    if not words:
        return []
    lines = [words[0]]
    for word in words[1:]:
        candidate = f"{lines[-1]} {word}"
        bbox = draw.textbbox((0, 0), candidate, font=font)
        if bbox[2] - bbox[0] <= max_width:
            lines[-1] = candidate
        else:
            lines.append(word)
    return lines


def draw_multiline(
    draw: ImageDraw.ImageDraw,
    x: int,
    y: int,
    text: str,
    font: ImageFont.ImageFont,
    fill: str,
    max_width: int,
    line_gap: int = 6,
    align: str = "left",
) -> int:
    lines = wrap_text(draw, text, font, max_width)
    cursor = y
    for line in lines:
        bbox = draw.textbbox((0, 0), line, font=font)
        width = bbox[2] - bbox[0]
        line_x = x if align == "left" else x - width // 2
        draw.text((line_x, cursor), line, font=font, fill=fill)
        cursor += (bbox[3] - bbox[1]) + line_gap
    return cursor - y - line_gap if lines else 0


def draw_blob(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], fill: str) -> None:
    x0, y0, x1, y1 = box
    draw.ellipse((x0 - 40, y0 - 40, x1 + 20, y1 + 30), fill=rgba(fill, 28))


def draw_sketch_box(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    fill: str,
    outline: str = LINE,
    radius: int = 20,
) -> None:
    x0, y0, x1, y1 = box
    draw.rounded_rectangle((x0 + 5, y0 + 6, x1 + 5, y1 + 6), radius=radius, fill=rgba(LINE, 18))
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=2)
    draw.rounded_rectangle((x0 + 2, y0 + 3, x1 - 1, y1 - 2), radius=radius, outline=outline, width=1)


def draw_hexagon(draw: ImageDraw.ImageDraw, center: tuple[int, int], width: int, height: int, fill: str, outline: str) -> tuple[int, int, int, int]:
    cx, cy = center
    hw = width // 2
    hh = height // 2
    points = [
        (cx - hw + 55, cy - hh),
        (cx + hw - 55, cy - hh),
        (cx + hw, cy),
        (cx + hw - 55, cy + hh),
        (cx - hw + 55, cy + hh),
        (cx - hw, cy),
    ]
    draw.polygon(points, fill=fill, outline=outline)
    return (cx - hw, cy - hh, cx + hw, cy + hh)


def draw_actor(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], label: str, icon: str, fill: str, fonts: dict[str, ImageFont.ImageFont]) -> None:
    draw_sketch_box(draw, box, fill)
    x0, y0, x1, y1 = box
    draw.rounded_rectangle((x0 + 12, y0 + 10, x0 + 68, y0 + 34), radius=10, fill=rgba("#FFFFFF", 170), outline=rgba(LINE, 90), width=1)
    draw.text((x0 + 24, y0 + 13), "ACTOR", font=fonts["micro"], fill=MUTED)
    draw.ellipse((x0 + 22, y0 + 42, x0 + 66, y0 + 86), fill=rgba("#FFFFFF", 180), outline=LINE, width=1)
    draw.text((x0 + 32, y0 + 46), icon, font=fonts["icon_small"], fill=TEXT)
    draw_multiline(draw, x0 + 16, y0 + 94, label, fonts["label"], TEXT, x1 - x0 - 32, align="left")


def draw_note(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    title: str,
    fill: str,
    fonts: dict[str, ImageFont.ImageFont],
    icon: str | None = None,
    reserve_icon_space: bool = False,
) -> None:
    draw_sketch_box(draw, box, fill, radius=14)
    x0, y0, x1, _ = box
    text_x = x0 + 16
    if icon:
        draw.ellipse((x0 + 12, y0 + 12, x0 + 42, y0 + 42), fill=rgba("#FFFFFF", 180), outline=LINE, width=1)
        draw.text((x0 + 19, y0 + 11), icon, font=fonts["note_bold"], fill=TEXT)
        text_x = x0 + 50
    elif reserve_icon_space:
        text_x = x0 + 50
    draw_multiline(draw, text_x, y0 + 14, title, fonts["note"], TEXT, x1 - text_x - 12)


def draw_adapter_badge(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], kind: str) -> None:
    x0, y0, x1, y1 = box
    draw.ellipse(box, fill=rgba("#FFFFFF", 180), outline=LINE, width=1)
    if kind == "rest":
        draw.line((x0 + 9, y0 + 14, x1 - 9, y0 + 14), fill=LINE, width=2)
        draw.line((x0 + 9, y0 + 22, x1 - 9, y0 + 22), fill=LINE, width=2)
        draw.line((x0 + 9, y0 + 30, x1 - 9, y0 + 30), fill=LINE, width=2)
    elif kind == "cli":
        draw.line((x0 + 10, y0 + 14, x0 + 18, y0 + 22, x0 + 10, y0 + 30), fill=LINE, width=2)
        draw.line((x0 + 22, y0 + 30, x1 - 10, y0 + 30), fill=LINE, width=2)
    elif kind == "inbound":
        draw.line((x0 + 8, y0 + 22, x1 - 12, y0 + 22), fill=LINE, width=2)
        draw.polygon([(x1 - 14, y0 + 16), (x1 - 4, y0 + 22), (x1 - 14, y0 + 28)], fill=LINE)
    elif kind == "outbound":
        draw.line((x1 - 8, y0 + 22, x0 + 12, y0 + 22), fill=LINE, width=2)
        draw.polygon([(x0 + 14, y0 + 16), (x0 + 4, y0 + 22), (x0 + 14, y0 + 28)], fill=LINE)
    elif kind == "postgres":
        draw_postgres_elephant(draw, (x0 + 4, y0 + 4, x1 - 4, y1 - 4))
    elif kind == "email":
        draw_envelope(draw, (x0 + 5, y0 + 8, x1 - 5, y1 - 6))
    elif kind == "rabbitmq":
        draw_rabbitmq(draw, (x0 + 4, y0 + 4, x1 - 4, y1 - 4))


def draw_mongo_leaf(draw: ImageDraw.ImageDraw, center: tuple[int, int], scale: float = 1.0) -> None:
    cx, cy = center
    leaf = [
        (cx, cy - 12 * scale),
        (cx + 8 * scale, cy - 2 * scale),
        (cx + 5 * scale, cy + 10 * scale),
        (cx, cy + 16 * scale),
        (cx - 5 * scale, cy + 10 * scale),
        (cx - 8 * scale, cy - 2 * scale),
    ]
    draw.polygon(leaf, fill="#2F855A", outline=LINE)
    draw.line((cx, cy - 10 * scale, cx, cy + 13 * scale), fill=rgba("#F7FAFC", 180), width=1)


def draw_postgres_elephant(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    cx = (x0 + x1) // 2
    cy = (y0 + y1) // 2
    draw.ellipse((cx - 18, cy - 16, cx + 18, cy + 14), fill="#4F78A8", outline=LINE)
    draw.ellipse((cx - 32, cy - 12, cx - 6, cy + 10), fill="#4F78A8", outline=LINE)
    draw.rectangle((cx - 30, cy + 8, cx - 18, cy + 26), fill="#4F78A8", outline=LINE)
    draw.ellipse((cx + 2, cy - 24, cx + 18, cy - 6), fill="#4F78A8", outline=LINE)
    draw.ellipse((cx - 16, cy - 26, cx, cy - 8), fill="#4F78A8", outline=LINE)
    draw.ellipse((cx - 18, cy - 5, cx - 12, cy + 1), fill="#FFFFFF")
    draw.ellipse((cx + 2, cy - 5, cx + 8, cy + 1), fill="#FFFFFF")


def draw_rabbitmq(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    draw.rounded_rectangle(box, radius=12, fill="#F97316", outline=LINE, width=1)
    cx = (x0 + x1) // 2
    cy = (y0 + y1) // 2 + 2
    draw.ellipse((cx - 12, cy - 8, cx + 12, cy + 12), fill="#FFF7ED", outline=LINE)
    draw.ellipse((cx - 11, cy - 24, cx - 3, cy - 6), fill="#FFF7ED", outline=LINE)
    draw.ellipse((cx + 3, cy - 24, cx + 11, cy - 6), fill="#FFF7ED", outline=LINE)
    draw.ellipse((cx - 6, cy, cx - 2, cy + 4), fill=LINE)
    draw.ellipse((cx + 2, cy, cx + 6, cy + 4), fill=LINE)
    draw.ellipse((cx - 2, cy + 6, cx + 2, cy + 10), fill="#F97316")


def draw_envelope(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    draw.rounded_rectangle(box, radius=8, fill="#FFFDF8", outline=LINE, width=2)
    draw.line((x0 + 4, y0 + 4, (x0 + x1) // 2, y0 + 16, x1 - 4, y0 + 4), fill=LINE, width=2)
    draw.line((x0 + 4, y1 - 4, (x0 + x1) // 2, y0 + 18, x1 - 4, y1 - 4), fill=LINE, width=2)


def draw_db_cylinder(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int]) -> None:
    x0, y0, x1, y1 = box
    draw.ellipse((x0, y0, x1, y0 + 16), fill="#F8FAFC", outline=LINE, width=2)
    draw.rectangle((x0, y0 + 8, x1, y1 - 8), fill="#F8FAFC", outline=LINE, width=2)
    draw.ellipse((x0, y1 - 16, x1, y1), fill="#F8FAFC", outline=LINE, width=2)


def draw_chip(draw: ImageDraw.ImageDraw, x: int, y: int, text: str, fill: str, fonts: dict[str, ImageFont.ImageFont]) -> int:
    text_box = draw.textbbox((0, 0), text, font=fonts["chip"])
    width = text_box[2] - text_box[0] + 28
    draw.rounded_rectangle((x, y, x + width, y + 34), radius=16, fill=fill, outline=LINE, width=1)
    draw.text((x + 14, y + 8), text, font=fonts["chip"], fill=TEXT)
    return width


def draw_annotation_chip(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    text: str,
    fonts: dict[str, ImageFont.ImageFont],
    fill: str = "#FFFDF8",
) -> None:
    draw.rounded_rectangle(box, radius=14, fill=fill, outline=rgba(LINE, 110), width=1)
    x0, y0, x1, y1 = box
    draw_multiline(draw, x0 + 12, y0 + 8, text, fonts["note"], MUTED, x1 - x0 - 24, line_gap=4)


def draw_poly_arrow(draw: ImageDraw.ImageDraw, points: list[tuple[int, int]], color: str, width: int = 3) -> None:
    draw.line(points, fill=color, width=width)
    (x0, y0), (x1, y1) = points[-2], points[-1]
    if abs(x1 - x0) >= abs(y1 - y0):
        if x1 >= x0:
            arrow = [(x1 - 14, y1 - 8), (x1 - 14, y1 + 8), (x1 + 1, y1)]
        else:
            arrow = [(x1 + 14, y1 - 8), (x1 + 14, y1 + 8), (x1 - 1, y1)]
    else:
        if y1 >= y0:
            arrow = [(x1 - 8, y1 - 14), (x1 + 8, y1 - 14), (x1, y1 + 1)]
        else:
            arrow = [(x1 - 8, y1 + 14), (x1 + 8, y1 + 14), (x1, y1 - 1)]
    draw.polygon(arrow, fill=color)


def draw_legend(draw: ImageDraw.ImageDraw, x: int, y: int, fonts: dict[str, ImageFont.ImageFont]) -> None:
    entries = [
        ("Drivers / Primary Actors", PALETTE["driver_actor"]),
        ("Driver / Primary Adapters", PALETTE["driver_adapter"]),
        ("Driver / Primary Ports", PALETTE["inbound"]),
        ("Core / Application", PALETTE["core"]),
        ("Driven / Secondary Ports", PALETTE["outbound"]),
        ("Driven / Secondary Adapters", PALETTE["driven_adapter"]),
    ]
    draw.text((x, y - 34), "Legend", font=fonts["chip"], fill=MUTED)
    for index, (label, fill) in enumerate(entries):
        col = index // 3
        row = index % 3
        draw_chip(draw, x + col * 286, y + row * 48, label, fill, fonts)


def draw_platform_box(
    draw: ImageDraw.ImageDraw,
    box: tuple[int, int, int, int],
    label: str,
    fonts: dict[str, ImageFont.ImageFont],
    icon_kind: str,
) -> None:
    draw_sketch_box(draw, box, "#F4F1EA", radius=14)
    x0, y0, x1, y1 = box
    icon_box = (x0 + 14, y0 + 14, x0 + 68, y0 + 68)
    draw.rounded_rectangle(icon_box, radius=12, fill="#FFFDF8", outline=LINE, width=1)
    if icon_kind == "postgres":
        draw_postgres_elephant(draw, (x0 + 18, y0 + 18, x0 + 64, y0 + 62))
    elif icon_kind == "rabbitmq":
        draw_rabbitmq(draw, (x0 + 18, y0 + 18, x0 + 64, y0 + 62))
    elif icon_kind == "email":
        draw_envelope(draw, (x0 + 18, y0 + 22, x0 + 64, y0 + 58))
    draw_multiline(draw, x0 + 82, y0 + 20, label, fonts["note_bold"], TEXT, x1 - x0 - 96, line_gap=4)


def draw_title_block(draw: ImageDraw.ImageDraw, fonts: dict[str, ImageFont.ImageFont]) -> None:
    draw.text((96, 58), "SYSTEM DESIGN / EDITORIAL FLOW", font=fonts["eyebrow"], fill=ACCENT)
    draw.text((96, 104), "Ports & Adapters", font=fonts["title"], fill=TEXT)
    draw.text((96, 170), "Hexagonal Architecture", font=fonts["title"], fill=TEXT)
    draw.rounded_rectangle((96, 244, 358, 284), radius=18, fill=rgba("#FFFFFF", 228), outline=rgba(PALETTE["accent_blue"], 140), width=1)
    draw.text((114, 255), "POLICY INSIDE / TECH OUTSIDE", font=fonts["chip"], fill=PALETTE["accent_blue"])
    draw_multiline(
        draw,
        96,
        310,
        "Business logic stays protected in the center. Ports define contracts. Adapters absorb transport and vendor detail so policy survives framework churn.",
        fonts["subtitle"],
        MUTED,
        930,
        line_gap=8,
    )


def build_image() -> Image.Image:
    image = Image.new("RGBA", (WIDTH, HEIGHT), BG)
    draw = ImageDraw.Draw(image, "RGBA")
    fonts = {
        "eyebrow": choose_font(FONT_SANS_BOLD, 26),
        "title": choose_font(FONT_SANS_BOLD, 56),
        "title_core": choose_font(FONT_SANS_BOLD, 50),
        "subtitle": choose_font(FONT_SANS, 26),
        "label": choose_font(FONT_SANS, 20),
        "note": choose_font(FONT_SANS, 16),
        "note_bold": choose_font(FONT_SANS_BOLD, 18),
        "chip": choose_font(FONT_SANS_BOLD, 16),
        "micro": choose_font(FONT_SANS_BOLD, 12),
        "icon": choose_font(FONT_SANS_BOLD, 40),
        "icon_small": choose_font(FONT_SANS_BOLD, 28),
        "footer": choose_font(FONT_SANS, 18),
    }

    draw_blob(draw, (-140, -140, 80, 70), ACCENT)
    draw_blob(draw, (1385, -100, 1570, 70), PALETTE["accent_blue"])
    draw_blob(draw, (1415, 996, 1596, 1184), ACCENT)
    draw_title_block(draw, fonts)

    user_box = (126, 358, 234, 514)
    job_box = (126, 568, 234, 724)
    rest_box = (282, 408, 450, 492)
    cli_box = (282, 618, 450, 702)
    inbound_box = (492, 466, 634, 576)
    outbound_top = (986, 446, 1132, 556)
    outbound_bottom = (986, 604, 1132, 714)
    mongo_box = (1180, 334, 1372, 418)
    email_box = (1180, 496, 1372, 580)
    queue_box = (1180, 658, 1372, 742)
    database_box = (1392, 324, 1554, 416)
    email_end_box = (1392, 486, 1554, 578)
    queue_end_box = (1392, 648, 1554, 740)

    draw_actor(draw, user_box, "User", "◔", PALETTE["driver_actor"], fonts)
    draw_actor(draw, job_box, "Batch Job /\nTest", "↻", PALETTE["driver_actor"], fonts)
    draw_note(draw, rest_box, "REST Adapter", PALETTE["driver_adapter"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (rest_box[0] + 12, rest_box[1] + 10, rest_box[0] + 42, rest_box[1] + 40), "rest")
    draw_note(draw, cli_box, "CLI / Test Adapter", PALETTE["driver_adapter"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (cli_box[0] + 12, cli_box[1] + 10, cli_box[0] + 42, cli_box[1] + 40), "cli")
    draw_note(draw, inbound_box, "Inbound Ports", PALETTE["inbound"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (inbound_box[0] + 12, inbound_box[1] + 10, inbound_box[0] + 42, inbound_box[1] + 40), "inbound")
    draw_note(draw, outbound_top, "Outbound Ports", PALETTE["outbound"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (outbound_top[0] + 12, outbound_top[1] + 10, outbound_top[0] + 42, outbound_top[1] + 40), "outbound")
    draw_note(draw, outbound_bottom, "Outbound Ports", PALETTE["outbound"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (outbound_bottom[0] + 12, outbound_bottom[1] + 10, outbound_bottom[0] + 42, outbound_bottom[1] + 40), "outbound")
    draw_note(draw, mongo_box, "PostgreSQL Adapter", PALETTE["driven_adapter"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (mongo_box[0] + 12, mongo_box[1] + 10, mongo_box[0] + 42, mongo_box[1] + 40), "postgres")
    draw_note(draw, email_box, "Email Adapter", PALETTE["driven_adapter"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (email_box[0] + 12, email_box[1] + 10, email_box[0] + 42, email_box[1] + 40), "email")
    draw_note(draw, queue_box, "RabbitMQ Adapter", PALETTE["driven_adapter"], fonts, reserve_icon_space=True)
    draw_adapter_badge(draw, (queue_box[0] + 12, queue_box[1] + 10, queue_box[0] + 42, queue_box[1] + 40), "rabbitmq")

    draw_platform_box(draw, database_box, "PostgreSQL", fonts, "postgres")
    draw_platform_box(draw, email_end_box, "Email Service", fonts, "email")
    draw_platform_box(draw, queue_end_box, "RabbitMQ", fonts, "rabbitmq")
    draw_hexagon(draw, (804, 540), 390, 256, PALETTE["core"], LINE)
    draw.rounded_rectangle((710, 432, 900, 464), radius=14, fill=rgba("#FFFFFF", 150), outline=rgba(LINE, 80), width=1)
    draw.text((742, 442), "POLICY BOUNDARY", font=fonts["chip"], fill=MUTED)
    draw_multiline(draw, 804, 484, "Business", fonts["title_core"], TEXT, 240, align="center")
    draw_multiline(draw, 804, 542, "Logic", fonts["title_core"], TEXT, 240, align="center")
    draw_multiline(draw, 804, 610, "(core application)", fonts["subtitle"], MUTED, 260, align="center")

    draw_poly_arrow(draw, [(234, 436), (282, 436)], LINE)
    draw_poly_arrow(draw, [(234, 646), (282, 646)], LINE)
    draw_annotation_chip(draw, (222, 378, 342, 414), "Actor intent", fonts)

    draw_poly_arrow(draw, [(450, 450), (492, 450), (492, 520)], LINE)
    draw_poly_arrow(draw, [(450, 660), (492, 660), (492, 544)], LINE)
    draw_annotation_chip(draw, (390, 374, 586, 418), "Adapter translates transport into use-case input", fonts)

    draw_poly_arrow(draw, [(634, 522), (674, 522)], LINE)
    draw_annotation_chip(draw, (630, 470, 782, 506), "Entry contract", fonts)

    draw_poly_arrow(draw, [(950, 498), (986, 498)], LINE)
    draw_poly_arrow(draw, [(950, 656), (986, 656)], LINE)
    draw_annotation_chip(draw, (874, 382, 1078, 426), "Core asks for capabilities", fonts)

    draw_poly_arrow(draw, [(1132, 500), (1180, 376)], LINE)
    draw_poly_arrow(draw, [(1132, 500), (1180, 538)], LINE)
    draw_poly_arrow(draw, [(1132, 658), (1180, 700)], LINE)
    draw_annotation_chip(draw, (1106, 392, 1318, 436), "Adapter binds vendor protocol", fonts)

    draw_poly_arrow(draw, [(1372, 376), (1392, 376)], LINE)
    draw_poly_arrow(draw, [(1372, 538), (1392, 538)], LINE)
    draw_poly_arrow(draw, [(1372, 700), (1392, 700)], LINE)

    draw_legend(draw, 354, 834, fonts)

    footer = "The shape is not the idea. The core owns policy. Ports define boundaries. Adapters absorb framework, transport, and vendor detail so business rules stay intact."
    draw_multiline(draw, 96, 1050, footer, fonts["footer"], MUTED, 1400, line_gap=5)
    return image.convert("RGB")


def main() -> None:
    parser = argparse.ArgumentParser(description="Render an editorial Ports & Adapters explainer.")
    parser.add_argument(
        "--output",
        default="assets/system-design/images/ports-adapters-hexagonal-architecture.png",
        help="Output PNG path",
    )
    args = parser.parse_args()
    output = Path(args.output)
    if not output.is_absolute():
        output = Path.cwd() / output
    output.parent.mkdir(parents=True, exist_ok=True)
    image = build_image()
    image.save(output, format="PNG")
    print(output)


if __name__ == "__main__":
    main()
