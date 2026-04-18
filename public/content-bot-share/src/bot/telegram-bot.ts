import TelegramBot from "node-telegram-bot-api";
import { config } from "../config.js";
import { db } from "../db.js";
import { logger } from "../utils/logger.js";

let bot: TelegramBot;

export function getBot(): TelegramBot {
  return bot;
}

function isAdmin(userId: number): boolean {
  if (config.ADMIN_IDS.length === 0) return true;
  return config.ADMIN_IDS.includes(userId);
}

// Escape ký tự đặc biệt của Telegram Markdown
function escMd(text: string): string {
  return text.replace(/[*_`\[\]()~>#+=|{}.!\\-]/g, "\\$&");
}

// Safe send — nếu Markdown lỗi thì gửi plain text
async function safeSend(chatId: number, text: string, markdown = false): Promise<void> {
  try {
    if (markdown) {
      await bot.sendMessage(chatId, text, { parse_mode: "Markdown", disable_web_page_preview: true });
    } else {
      await bot.sendMessage(chatId, text, { disable_web_page_preview: true });
    }
  } catch (err: any) {
    if (err.message?.includes("parse entities") || err.message?.includes("Bad Request")) {
      // Retry without Markdown
      const plain = text.replace(/[*_`\[\]]/g, "");
      await bot.sendMessage(chatId, plain, { disable_web_page_preview: true });
    } else {
      logger.error("bot", `Send message failed: ${err.message}`);
    }
  }
}

export function startBot(): TelegramBot {
  bot = new TelegramBot(config.BOT_TOKEN, { polling: true });

  // Catch polling errors to prevent crash
  bot.on("polling_error", (err: any) => {
    logger.error("bot", `Polling error: ${err.code || err.message}`);
  });

  logger.info("bot", "Telegram bot started");

  // /start
  bot.onText(/\/start/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      await safeSend(
        msg.chat.id,
        `🤖 Content Bot đã sẵn sàng!\n\n` +
          `📥 /add <text> - Thêm nội dung vào Rổ Content\n` +
          `📊 /status - Xem tổng quan\n` +
          `📋 /queue - Xem hàng đợi\n` +
          `📰 /recent - Bài đã đăng gần đây\n` +
          `🔗 /sources - Danh sách nguồn\n` +
          `➕ /addsource <type> <handle> <name>\n` +
          `➖ /removesource <handle>\n` +
          `🔄 /retry - Retry các item lỗi\n` +
          `⏸ /pause - Tạm dừng auto-publish\n` +
          `▶️ /resume - Bật lại auto-publish\n` +
          `🚀 /crawlnow - Crawl ngay\n` +
          `📝 /logs - Xem log gần đây\n\n` +
          `💡 Gửi tin nhắn bất kỳ (không có /) sẽ tự động thêm vào Rổ Content.`
      );
    } catch (err: any) {
      logger.error("bot", `Start cmd error: ${err.message}`);
    }
  });

  // /status
  bot.onText(/\/status/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const today = new Date();
      today.setHours(0, 0, 0, 0);

      const [pending, processing, rewritten, published, failed, todayPublished] =
        await Promise.all([
          db.contentItem.count({ where: { status: "pending" } }),
          db.contentItem.count({ where: { status: "processing" } }),
          db.contentItem.count({ where: { status: "rewritten" } }),
          db.contentItem.count({ where: { status: "published" } }),
          db.contentItem.count({ where: { status: "failed" } }),
          db.contentItem.count({
            where: { status: "published", publishedAt: { gte: today } },
          }),
        ]);

      const autoPub = await getSettingValue("auto_publish", "true");
      const sources = await db.source.count({ where: { isActive: true } });

      await safeSend(
        msg.chat.id,
        `📊 Tổng Quan\n\n` +
          `⏳ Pending: ${pending}\n` +
          `🔄 Processing: ${processing}\n` +
          `✍️ Rewritten: ${rewritten}\n` +
          `✅ Published: ${published} (hôm nay: ${todayPublished})\n` +
          `❌ Failed: ${failed}\n\n` +
          `🔗 Sources: ${sources}\n` +
          `📡 Auto-publish: ${autoPub === "true" ? "BẬT" : "TẮT"}\n` +
          `🤖 DeepSeek: ${config.hasDeepSeek ? "OK" : "Chưa cấu hình"}\n` +
          `📱 TG Crawler: ${config.hasTelegramCrawler ? "OK" : "Chưa cấu hình"}\n` +
          `🐦 Twitter: ${config.hasTwitter ? "OK" : "Chưa cấu hình"}`
      );
    } catch (err: any) {
      logger.error("bot", `Status cmd error: ${err.message}`);
    }
  });

  // /add <text>
  bot.onText(/\/add (.+)/, async (msg, match) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const text = match![1].trim();
      if (!text) {
        await safeSend(msg.chat.id, "⚠️ Vui lòng nhập nội dung: /add <text>");
        return;
      }

      await db.contentItem.create({
        data: {
          originalText: text,
          externalId: `manual_${Date.now()}_${msg.message_id}`,
          authorName: msg.from?.first_name || "Manual",
          status: "pending",
        },
      });

      await safeSend(msg.chat.id, `✅ Đã thêm vào Rổ Content!\n📝 "${text.slice(0, 100)}${text.length > 100 ? "..." : ""}"`);
      logger.info("bot", `Manual content added by ${msg.from?.first_name}`);
    } catch (err: any) {
      logger.error("bot", `Add cmd error: ${err.message}`);
    }
  });

  // /sources
  bot.onText(/\/sources/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const sources = await db.source.findMany({
        orderBy: { createdAt: "desc" },
      });

      if (sources.length === 0) {
        await safeSend(msg.chat.id, "📭 Chưa có nguồn nào.\nThêm: /addsource <twitter|telegram> <handle> <tên>");
        return;
      }

      const lines = sources.map(
        (s) => `${s.isActive ? "🟢" : "🔴"} [${s.type}] ${s.name} (${s.handle})`
      );

      await safeSend(msg.chat.id, `🔗 Danh sách nguồn:\n\n${lines.join("\n")}`);
    } catch (err: any) {
      logger.error("bot", `Sources cmd error: ${err.message}`);
    }
  });

  // /addsource <type> <handle> <name...>
  bot.onText(/\/addsource (\S+) (\S+) (.+)/, async (msg, match) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const type = match![1].toLowerCase();
      const handle = match![2];
      const name = match![3].trim();

      if (!["twitter", "telegram"].includes(type)) {
        await safeSend(msg.chat.id, "⚠️ Type phải là: twitter hoặc telegram");
        return;
      }

      await db.source.create({
        data: { type, handle, name },
      });
      await safeSend(msg.chat.id, `✅ Đã thêm nguồn: [${type}] ${name} (${handle})`);
      logger.info("bot", `Source added: ${type} ${handle}`);
    } catch (e: any) {
      if (e.code === "P2002") {
        await safeSend(msg.chat.id, `⚠️ Nguồn ${match![2]} đã tồn tại!`);
      } else {
        await safeSend(msg.chat.id, `❌ Lỗi: ${e.message}`);
      }
    }
  });

  // /removesource <handle>
  bot.onText(/\/removesource (\S+)/, async (msg, match) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const handle = match![1];

      const source = await db.source.findFirst({ where: { handle } });
      if (!source) {
        await safeSend(msg.chat.id, `⚠️ Không tìm thấy nguồn: ${handle}`);
        return;
      }

      await db.source.update({
        where: { id: source.id },
        data: { isActive: false },
      });

      await safeSend(msg.chat.id, `🔴 Đã tắt nguồn: ${source.name} (${handle})`);
      logger.info("bot", `Source deactivated: ${handle}`);
    } catch (err: any) {
      logger.error("bot", `Removesource cmd error: ${err.message}`);
    }
  });

  // /queue
  bot.onText(/\/queue/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const items = await db.contentItem.findMany({
        where: { status: "pending" },
        orderBy: { crawledAt: "desc" },
        take: 10,
        include: { source: true },
      });

      if (items.length === 0) {
        await safeSend(msg.chat.id, "📭 Không có item nào trong hàng đợi.");
        return;
      }

      // Escape nội dung gốc để tránh lỗi Markdown
      const lines = items.map(
        (item, i) => {
          const preview = item.originalText.slice(0, 80).replace(/[*_`\[\]]/g, "");
          return `${i + 1}. ${preview}...\n   📌 ${item.source?.name || "Manual"} | ${item.id.slice(0, 8)}`;
        }
      );

      await safeSend(msg.chat.id, `📋 Hàng đợi (${items.length}):\n\n${lines.join("\n\n")}`);
    } catch (err: any) {
      logger.error("bot", `Queue cmd error: ${err.message}`);
    }
  });

  // /recent
  bot.onText(/\/recent/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const items = await db.contentItem.findMany({
        where: { status: "published" },
        orderBy: { publishedAt: "desc" },
        take: 10,
      });

      if (items.length === 0) {
        await safeSend(msg.chat.id, "📭 Chưa có bài nào được đăng.");
        return;
      }

      const lines = items.map(
        (item, i) => {
          const preview = (item.rewrittenText || item.originalText).slice(0, 80).replace(/[*_`\[\]]/g, "");
          return `${i + 1}. ${preview}...\n   🕐 ${item.publishedAt?.toLocaleString("vi-VN")}`;
        }
      );

      await safeSend(msg.chat.id, `📰 Đã đăng gần đây:\n\n${lines.join("\n\n")}`);
    } catch (err: any) {
      logger.error("bot", `Recent cmd error: ${err.message}`);
    }
  });

  // /retry
  bot.onText(/\/retry/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const result = await db.contentItem.updateMany({
        where: { status: "failed" },
        data: { status: "pending", failReason: null },
      });

      await safeSend(msg.chat.id, `🔄 Đã retry ${result.count} items.`);
      logger.info("bot", `Retried ${result.count} failed items`);
    } catch (err: any) {
      logger.error("bot", `Retry cmd error: ${err.message}`);
    }
  });

  // /pause
  bot.onText(/\/pause/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      await setSetting("auto_publish", "false");
      await safeSend(msg.chat.id, "⏸ Auto-publish đã TẮT. Nội dung sẽ dừng ở trạng thái 'rewritten'.");
      logger.info("bot", "Auto-publish paused");
    } catch (err: any) {
      logger.error("bot", `Pause cmd error: ${err.message}`);
    }
  });

  // /resume
  bot.onText(/\/resume/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      await setSetting("auto_publish", "true");
      await safeSend(msg.chat.id, "▶️ Auto-publish đã BẬT.");
      logger.info("bot", "Auto-publish resumed");
    } catch (err: any) {
      logger.error("bot", `Resume cmd error: ${err.message}`);
    }
  });

  // /logs
  bot.onText(/\/logs/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const logs = await db.log.findMany({
        orderBy: { createdAt: "desc" },
        take: 20,
      });

      if (logs.length === 0) {
        await safeSend(msg.chat.id, "📭 Chưa có log.");
        return;
      }

      const lines = logs.map(
        (l) => `[${l.level.toUpperCase()}] ${l.module}: ${l.message.slice(0, 60)}`
      );

      await safeSend(msg.chat.id, `📝 Logs:\n\n${lines.join("\n")}`);
    } catch (err: any) {
      logger.error("bot", `Logs cmd error: ${err.message}`);
    }
  });

  // /skip <id>
  bot.onText(/\/skip (\S+)/, async (msg, match) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      const idPrefix = match![1];

      const item = await db.contentItem.findFirst({
        where: { id: { startsWith: idPrefix } },
      });

      if (!item) {
        await safeSend(msg.chat.id, `⚠️ Không tìm thấy item: ${idPrefix}`);
        return;
      }

      await db.contentItem.update({
        where: { id: item.id },
        data: { status: "skipped" },
      });

      await safeSend(msg.chat.id, `⏭ Đã bỏ qua item ${item.id.slice(0, 8)}`);
    } catch (err: any) {
      logger.error("bot", `Skip cmd error: ${err.message}`);
    }
  });

  // /crawlnow
  bot.onText(/\/crawlnow/, async (msg) => {
    if (!isAdmin(msg.from!.id)) return;
    try {
      await safeSend(msg.chat.id, "🚀 Đang crawl...");
      bot.emit("crawl_now", msg.chat.id);
    } catch (err: any) {
      logger.error("bot", `Crawlnow cmd error: ${err.message}`);
    }
  });

  // Default: any text without command prefix → manual input
  bot.on("message", async (msg) => {
    if (!msg.text || msg.text.startsWith("/")) return;
    if (!isAdmin(msg.from!.id)) return;

    try {
      await db.contentItem.create({
        data: {
          originalText: msg.text,
          externalId: `manual_${Date.now()}_${msg.message_id}`,
          authorName: msg.from?.first_name || "Manual",
          status: "pending",
        },
      });

      await safeSend(
        msg.chat.id,
        `✅ Đã thêm vào Rổ Content!\n📝 "${msg.text.slice(0, 80)}${msg.text.length > 80 ? "..." : ""}"`
      );
    } catch (err: any) {
      logger.error("bot", `Message handler error: ${err.message}`);
    }
  });

  return bot;
}

// Settings helpers
export async function getSettingValue(key: string, defaultValue: string): Promise<string> {
  const setting = await db.setting.findUnique({ where: { key } });
  return setting?.value ?? defaultValue;
}

export async function setSetting(key: string, value: string): Promise<void> {
  await db.setting.upsert({
    where: { key },
    update: { value },
    create: { key, value },
  });
}
