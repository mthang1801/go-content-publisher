import OpenAI from "openai";
import { config } from "../config.js";
import { db } from "../db.js";
import { logger } from "../utils/logger.js";

let client: OpenAI | null = null;

function getClient(): OpenAI {
  if (!client) {
    client = new OpenAI({
      apiKey: config.DEEPSEEK_API_KEY,
      baseURL: "https://api.deepseek.com",
    });
  }
  return client;
}

const SYSTEM_PROMPT = `Bạn là một biên tập viên tin tức tài chính - kinh tế - chính trị quốc tế chuyên nghiệp. Nhiệm vụ:

1. VIẾT LẠI nội dung gốc thành bài tin tức ngắn gọn, khách quan, chuyên nghiệp **BẰNG TIẾNG VIỆT**
2. KIỂM CHỨNG: đánh giá độ tin cậy của thông tin
3. Quyết định có nên đăng hay không

GÓC NHÌN KHÁCH QUAN - TOÀN DIỆN:
- Viết cho độc giả quan tâm đến THỊ TRƯỜNG TÀI CHÍNH TOÀN CẦU (chứng khoán, forex, hàng hóa, crypto, trái phiếu...)
- Bao gồm cả KINH TẾ VĨ MÔ (GDP, lạm phát, lãi suất, chính sách tiền tệ, thương mại...)
- Bao gồm cả CHÍNH TRỊ - ĐỊA CHÍNH TRỊ (quan hệ quốc tế, xung đột, bầu cử, chính sách...)
- KHÔNG thiên lệch về crypto - nếu tin liên quan crypto thì viết như một phần của thị trường tài chính chung
- Phân tích tác động đa chiều: ảnh hưởng đến nhiều loại tài sản, nhiều thị trường, không chỉ crypto
- Giọng văn trung lập, khách quan như hãng tin tài chính quốc tế (Bloomberg, Reuters)

QUAN TRỌNG - NGÔN NGỮ:
- Nội dung đầu vào có thể ở BẤT KỲ ngôn ngữ nào (English, 中文, 日本語, 한국어, Русский, v.v.)
- BẮT BUỘC viết lại TOÀN BỘ bằng TIẾNG VIỆT
- Giữ nguyên tên riêng: tên người, tên tổ chức, tên dự án, tên token/coin
- Giữ nguyên các thuật ngữ tài chính/crypto phổ biến không cần dịch (Fed, CPI, GDP, DeFi, NFT, DEX, ETF, S&P 500, Nasdaq...)
- Các số liệu, %, giá trị USD giữ nguyên

LỌC NỘI DUNG QUẢNG CÁO / SPAM (ƯU TIÊN CAO NHẤT):
Trước khi viết lại, PHẢI kiểm tra nội dung gốc có phải quảng cáo/spam không. Nếu phát hiện BẤT KỲ dấu hiệu nào sau đây → set shouldPublish=false, KHÔNG viết lại:
- Quảng cáo sản phẩm, dịch vụ, sàn giao dịch (VD: "Đăng ký ngay", "Mở tài khoản", "Nhận bonus")
- Chèn link referral, affiliate, mã giới thiệu, mã giảm giá
- Link mời vào group, channel khác (t.me/..., discord.gg/..., bit.ly/...)
- Kêu gọi đầu tư, hứa lợi nhuận ("x100", "guaranteed profit", "free money")
- Shill token/coin cụ thể với mục đích pump ("buy now", "moon soon", "gem x1000")
- Nội dung airdrop yêu cầu làm task (follow, like, retweet, join)
- Bài PR/sponsored có chứa liên kết tracking hoặc UTM parameters
- Nội dung copy-paste quảng cáo rõ ràng, thiếu giá trị tin tức

Quy tắc viết (chỉ áp dụng nếu nội dung KHÔNG phải quảng cáo):
- BẮT BUỘC VIẾT LẠI HOÀN TOÀN - KHÔNG BAO GIỜ copy nguyên văn nội dung gốc
- Phải diễn đạt lại bằng câu từ MỚI, cấu trúc câu KHÁC, góc nhìn biên tập viên
- Giữ nguyên số liệu, tên người, tên dự án nhưng PHẢI viết lại câu văn
- Thêm context/giải thích nếu cần cho độc giả Việt Nam
- Bài viết lại nên từ 100-300 từ
- Nếu nội dung gốc chứa nhiều đoạn tin (phân tách bởi ---), hãy TỔNG HỢP thành 1 bài viết duy nhất, lọc bỏ thông tin trùng lặp, giữ các điểm chính
- Format Markdown cho Telegram (*bold*, _italic_)
- KHÔNG thêm thông tin sai hoặc phóng đại
- Giọng văn chuyên nghiệp, khách quan, trung lập như biên tập viên hãng tin tài chính quốc tế
- LOẠI BỎ tất cả link, URL, liên kết (http, https, t.me, bit.ly...) khỏi bài viết lại
- LOẠI BỎ tên kênh nguồn, tên channel, tên tài khoản Twitter nguồn
- Chỉ giữ nội dung tin tức thuần túy, KHÔNG gắn nguồn

Trả lời dạng JSON:
{
  "rewrittenText": "Nội dung đã viết lại BẰNG TIẾNG VIỆT (Markdown format cho Telegram). Để trống nếu shouldPublish=false",
  "rewrittenTextEn": "BẢN DỊCH TIẾNG ANH của bài viết tiếng Việt ở trên (cùng nội dung, professional English, Markdown format). Để trống nếu shouldPublish=false",
  "tweetVI": "Phiên bản SIÊU NGẮN GỌN bằng TIẾNG VIỆT cho Twitter (tối đa 4 dòng, dưới 250 ký tự). Chỉ giữ thông tin cốt lõi nhất. Không dùng Markdown. Để trống nếu shouldPublish=false",
  "tweetEN": "Phiên bản SIÊU NGẮN GỌN bằng TIẾNG ANH cho Twitter (max 4 lines, under 250 chars). Core info only. No Markdown. Empty if shouldPublish=false",
  "factCheckNote": "Ghi chú kiểm chứng bằng tiếng Việt",
  "shouldPublish": true/false,
  "reason": "Lý do nếu không nên đăng - VD: 'Nội dung quảng cáo sàn X', 'Chứa link referral', 'Shill token không có giá trị tin tức'"
}`;

interface ProcessResult {
  rewrittenText: string;
  rewrittenTextEn: string;
  tweetVI: string;
  tweetEN: string;
  factCheckNote: string;
  shouldPublish: boolean;
  reason?: string;
}

async function callDeepSeek(originalText: string): Promise<ProcessResult> {
  const client = getClient();

  const response = await client.chat.completions.create({
    model: config.DEEPSEEK_MODEL,
    messages: [
      { role: "system", content: SYSTEM_PROMPT },
      {
        role: "user",
        content: `Nội dung gốc:\n\n${originalText}`,
      },
    ],
    temperature: 0.7,
    max_tokens: 2048,
    response_format: { type: "json_object" },
  });

  const text = response.choices[0]?.message?.content || "";

  try {
    const parsed = JSON.parse(text);
    return {
      rewrittenText: parsed.rewrittenText || "",
      rewrittenTextEn: parsed.rewrittenTextEn || "",
      tweetVI: parsed.tweetVI || "",
      tweetEN: parsed.tweetEN || "",
      factCheckNote: parsed.factCheckNote || "",
      shouldPublish: parsed.shouldPublish !== false,
      reason: parsed.reason,
    };
  } catch {
    // Fallback: try to extract JSON from text
    const jsonMatch = text.match(/\{[\s\S]*\}/);
    if (jsonMatch) {
      const parsed = JSON.parse(jsonMatch[0]);
      return {
        rewrittenText: parsed.rewrittenText || "",
        rewrittenTextEn: parsed.rewrittenTextEn || "",
        factCheckNote: parsed.factCheckNote || "",
        shouldPublish: parsed.shouldPublish !== false,
        reason: parsed.reason,
      };
    }
    throw new Error("Cannot parse DeepSeek response as JSON");
  }
}

// ===== DUPLICATE DETECTION =====
// Tính similarity giữa 2 chuỗi dùng bigram (2-gram) overlap
function bigrams(text: string): Set<string> {
  const normalized = text.toLowerCase().replace(/[^\p{L}\p{N}\s]/gu, "").replace(/\s+/g, " ").trim();
  const set = new Set<string>();
  for (let i = 0; i < normalized.length - 1; i++) {
    set.add(normalized.slice(i, i + 2));
  }
  return set;
}

function similarity(textA: string, textB: string): number {
  const a = bigrams(textA);
  const b = bigrams(textB);
  if (a.size === 0 || b.size === 0) return 0;
  let intersection = 0;
  for (const gram of a) {
    if (b.has(gram)) intersection++;
  }
  return (2 * intersection) / (a.size + b.size);
}

const DUPLICATE_THRESHOLD = 0.7; // 70% trùng lặp → bỏ qua

async function isDuplicate(originalText: string): Promise<string | null> {
  // So sánh với các bài đã xử lý trong 48h gần đây
  const since = new Date(Date.now() - 48 * 60 * 60 * 1000);
  const recentItems = await db.contentItem.findMany({
    where: {
      status: { in: ["rewritten", "published", "processing", "skipped"] },
      crawledAt: { gte: since },
    },
    select: { id: true, originalText: true, rewrittenText: true },
  });

  for (const existing of recentItems) {
    // So sánh originalText
    const simOrig = similarity(originalText, existing.originalText);
    if (simOrig >= DUPLICATE_THRESHOLD) {
      return `Trùng lặp ${Math.round(simOrig * 100)}% với item ${existing.id.slice(0, 8)}`;
    }

    // So sánh với rewrittenText (bắt trùng cross-source: cùng tin từ nguồn khác)
    if (existing.rewrittenText) {
      const simRewrite = similarity(originalText, existing.rewrittenText);
      if (simRewrite >= 0.65) { // Cross-language check
        return `Trùng nội dung ${Math.round(simRewrite * 100)}% với item ${existing.id.slice(0, 8)}`;
      }
    }
  }
  return null;
}

// Check duplicate SAU KHI viết lại — so sánh rewrittenText mới với các bài đã published
async function isDuplicateRewritten(rewrittenText: string): Promise<string | null> {
  const since = new Date(Date.now() - 48 * 60 * 60 * 1000);
  const recentItems = await db.contentItem.findMany({
    where: {
      status: { in: ["rewritten", "published"] },
      rewrittenText: { not: null },
      crawledAt: { gte: since },
    },
    select: { id: true, rewrittenText: true },
  });

  for (const existing of recentItems) {
    if (!existing.rewrittenText) continue;
    const sim = similarity(rewrittenText, existing.rewrittenText);
    if (sim >= 0.75) { // 75% trùng bài viết lại → cùng chủ đề
      return `Bài viết lại trùng ${Math.round(sim * 100)}% với item ${existing.id.slice(0, 8)}`;
    }
  }
  return null;
}

// ===== MAIN PROCESSOR =====
export async function processNextPending(): Promise<boolean> {
  if (!config.hasDeepSeek) return false;

  // Auto-recover stuck "processing" items (older than 5 min)
  const stuckCutoff = new Date(Date.now() - 5 * 60 * 1000);
  await db.contentItem.updateMany({
    where: { status: "processing", crawledAt: { lt: stuckCutoff } },
    data: { status: "pending" },
  });

  // Pick one pending item
  const item = await db.contentItem.findFirst({
    where: { status: "pending" },
    orderBy: { crawledAt: "asc" },
  });

  if (!item) return false;

  // Check duplicate BEFORE calling DeepSeek (save API cost)
  const dupReason = await isDuplicate(item.originalText);
  if (dupReason) {
    await db.contentItem.update({
      where: { id: item.id },
      data: {
        status: "skipped",
        failReason: dupReason,
      },
    });
    logger.info("deepseek", `Item ${item.id.slice(0, 8)} skipped: ${dupReason}`);
    return true;
  }

  // Mark as processing
  await db.contentItem.update({
    where: { id: item.id },
    data: { status: "processing" },
  });

  try {
    const result = await callDeepSeek(item.originalText);

    if (result.shouldPublish && result.rewrittenText) {
      // Check duplicate SAU KHI viết lại (bắt trùng cross-source)
      const dupRewritten = await isDuplicateRewritten(result.rewrittenText);
      if (dupRewritten) {
        await db.contentItem.update({
          where: { id: item.id },
          data: {
            status: "skipped",
            rewrittenText: result.rewrittenText,
            failReason: dupRewritten,
          },
        });
        logger.info("deepseek", `Item ${item.id.slice(0, 8)} skipped post-rewrite: ${dupRewritten}`);
        return true;
      }

      await db.contentItem.update({
        where: { id: item.id },
        data: {
          status: "rewritten",
          rewrittenText: result.rewrittenText,
          rewrittenTextEn: result.rewrittenTextEn || null,
          tweetTextVI: result.tweetVI || null,
          tweetTextEN: result.tweetEN || null,
          factCheckNote: result.factCheckNote,
        },
      });
      logger.info("deepseek", `Processed item ${item.id.slice(0, 8)} → rewritten`);
    } else {
      await db.contentItem.update({
        where: { id: item.id },
        data: {
          status: "failed",
          factCheckNote: result.factCheckNote,
          failReason: result.reason || "Fact-check flagged: should not publish",
        },
      });
      logger.warn("deepseek", `Item ${item.id.slice(0, 8)} flagged: ${result.reason}`);
    }

    return true;
  } catch (err: any) {
    await db.contentItem.update({
      where: { id: item.id },
      data: {
        status: "failed",
        failReason: `DeepSeek error: ${err.message}`,
      },
    });
    logger.error("deepseek", `Error processing ${item.id.slice(0, 8)}: ${err.message}`);
    return false;
  }
}
