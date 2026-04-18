import { config } from "../config.js";
import { db } from "../db.js";
import { logger } from "../utils/logger.js";
import { getBot, getSettingValue } from "../bot/telegram-bot.js";
import { publishToTwitterVI, publishToTwitterEN } from "./twitter-publisher.js";

export async function publishNextReady(): Promise<boolean> {
  // Check auto-publish setting
  const autoPub = await getSettingValue("auto_publish", config.AUTO_PUBLISH ? "true" : "false");
  if (autoPub !== "true") return false;

  // Pick one rewritten item
  const item = await db.contentItem.findFirst({
    where: { status: "rewritten" },
    orderBy: { crawledAt: "asc" },
  });

  if (!item) return false;

  const bot = getBot();
  if (!bot) {
    logger.error("publisher", "Bot not initialized, cannot publish");
    return false;
  }

  try {
    // Only publish rewritten content, never original
    if (!item.rewrittenText) {
      logger.warn("publisher", `Item ${item.id.slice(0, 8)} has no rewritten text, skipping`);
      return false;
    }

    const textVI = stripAllSourceInfo(item.rewrittenText);

    // Skip if after stripping, content is too short
    if (textVI.length < 20) {
      logger.warn("publisher", `Item ${item.id.slice(0, 8)} too short after strip, skipping`);
      await db.contentItem.update({
        where: { id: item.id },
        data: { status: "skipped", failReason: "Content too short after stripping sources" },
      });
      return false;
    }

    // === 1. Publish to Telegram Channel ===
    let sent;
    try {
      sent = await bot.sendMessage(config.TARGET_CHANNEL, textVI, {
        parse_mode: "Markdown",
        disable_web_page_preview: true,
      });
    } catch (mdErr: any) {
      // Markdown parse error → retry without parse_mode
      if (mdErr.message?.includes("parse entities") || mdErr.message?.includes("Bad Request")) {
        const plainText = textVI.replace(/[*_`\[\]]/g, "");
        sent = await bot.sendMessage(config.TARGET_CHANNEL, plainText, {
          disable_web_page_preview: true,
        });
        logger.warn("publisher", `Item ${item.id.slice(0, 8)} sent as plain text (Markdown error)`);
      } else {
        throw mdErr;
      }
    }

    // Mark as published immediately (Twitter will publish async)
    await db.contentItem.update({
      where: { id: item.id },
      data: {
        status: "published",
        publishedAt: new Date(),
        publishedMsgId: sent.message_id.toString(),
      },
    });

    logger.info("publisher", `Published ${item.id.slice(0, 8)} → [TG]`);
    return true;
  } catch (err: any) {
    await db.contentItem.update({
      where: { id: item.id },
      data: {
        status: "failed",
        failReason: `Publish error: ${err.message}`,
      },
    });
    logger.error("publisher", `Failed to publish ${item.id.slice(0, 8)}: ${err.message}`);
    return false;
  }
}

// === Twitter publish riêng, chạy chậm hơn ===
export async function publishNextToTwitter(): Promise<boolean> {
  // Tìm item đã published lên TG nhưng chưa tweet
  const item = await db.contentItem.findFirst({
    where: {
      status: "published",
      tweetViId: null,
      tweetEnId: null,
      rewrittenText: { not: null },
    },
    orderBy: { publishedAt: "asc" },
  });

  if (!item) return false;

  try {
    let tweetViId: string | null = null;
    let tweetEnId: string | null = null;

    // Tweet VI - ưu tiên dùng bản ngắn từ DeepSeek
    if (config.hasTwitterPublishVI) {
      const textVI = item.tweetTextVI
        ? stripAllSourceInfo(item.tweetTextVI)
        : stripAllSourceInfo(item.rewrittenText!);
      if (textVI.length >= 20) {
        tweetViId = await publishToTwitterVI(textVI);
      }
    }

    // Tweet EN - ưu tiên dùng bản ngắn từ DeepSeek
    if (config.hasTwitterPublishEN) {
      const textEN = item.tweetTextEN
        ? stripAllSourceInfo(item.tweetTextEN)
        : item.rewrittenTextEn
          ? stripAllSourceInfo(item.rewrittenTextEn)
          : null;
      if (textEN && textEN.length >= 20) {
        tweetEnId = await publishToTwitterEN(textEN);
      }
    }

    await db.contentItem.update({
      where: { id: item.id },
      data: {
        tweetViId: tweetViId || "skipped",
        tweetEnId: tweetEnId || "skipped",
      },
    });

    const platforms: string[] = [];
    if (tweetViId && tweetViId !== "skipped") platforms.push("TW-VI");
    if (tweetEnId && tweetEnId !== "skipped") platforms.push("TW-EN");

    if (platforms.length > 0) {
      logger.info("twitter-pub", `Tweeted ${item.id.slice(0, 8)} → [${platforms.join(", ")}]`);
    }
    return true;
  } catch (err: any) {
    logger.error("twitter-pub", `Tweet failed ${item.id.slice(0, 8)}: ${err.message}`);
    // Mark as skipped to avoid retry loop
    await db.contentItem.update({
      where: { id: item.id },
      data: { tweetViId: "failed", tweetEnId: "failed" },
    });
    return false;
  }
}

// Lọc triệt để mọi link, nguồn, attribution khỏi bài viết
function stripAllSourceInfo(text: string): string {
  let cleaned = text;

  // 1. Loại bỏ Markdown link [text](url) → giữ text
  cleaned = cleaned.replace(/\[([^\]]*)\]\([^)]*\)/g, "$1");

  // 2. Loại bỏ mọi URL dạng http/https
  cleaned = cleaned.replace(/https?:\/\/[^\s)>\]]+/gi, "");

  // 3. Loại bỏ mọi link t.me, bit.ly, discord.gg, tinyurl...
  cleaned = cleaned.replace(/(t\.me|bit\.ly|discord\.gg|tinyurl\.com|goo\.gl)\/[^\s)>\]]+/gi, "");

  // 4. Loại bỏ @username handles
  cleaned = cleaned.replace(/@\w+/g, "");

  // 5. Loại bỏ dòng chứa từ khóa nguồn
  cleaned = cleaned.replace(/^.*?(Nguồn|Source|Theo|Via|Từ|Credit|Tham khảo|Xem thêm|Chi tiết tại|Đọc thêm)\s*[:：].*/gim, "");

  // 6. Loại bỏ dòng chỉ là tên nguồn phổ biến
  const sourceNames = [
    "VN Wallstreet", "UG Wallstreet", "Wu Blockchain", "Shauotat Official", "Clash Report",
    "CoinDesk", "CoinTelegraph", "The Block", "Bloomberg", "Reuters",
    "VN Wall Street", "Sha ướt át",
  ];
  for (const name of sourceNames) {
    const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    cleaned = cleaned.replace(new RegExp(`^.*${escaped}.*$`, "gim"), "");
  }

  // 7. Loại bỏ dòng bắt đầu bằng emoji phổ biến dùng cho nguồn/ghi chú
  cleaned = cleaned.replace(/^[\s]*[\u{1F4CC}\u{1F4CE}\u{1F517}\u{1F4A1}\u{1F4F0}\u{1F4E2}\u{1F50D}\u{2139}\u{26A0}\u{270F}].*$/gmu, "");

  // 8. Loại bỏ dòng "Thông tin dựa trên...", "Lưu ý:...", "Khuyến nghị..."
  cleaned = cleaned.replace(/^.*?(Thông tin dựa trên|Lưu ý|Khuyến nghị|Disclaimer|Note|Cần lưu ý|Khuyến cáo).*$/gim, "");

  // 9. Loại bỏ dòng Telegram link preview (chứa "Telegram" + tên channel)
  cleaned = cleaned.replace(/^.*?Telegram\s*\n.*?$/gim, "");

  // 10. Loại bỏ "VIEW MESSAGE" text
  cleaned = cleaned.replace(/VIEW MESSAGE/gi, "");

  // 11. Dọn dẹp: dòng trống liên tiếp, khoảng trắng thừa
  cleaned = cleaned.replace(/\n{3,}/g, "\n\n");
  cleaned = cleaned.replace(/[ \t]+$/gm, "");

  return cleaned.trim();
}
