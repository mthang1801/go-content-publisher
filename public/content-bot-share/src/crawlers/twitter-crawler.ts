import { TwitterApi } from "twitter-api-v2";
import { config } from "../config.js";
import { db } from "../db.js";
import { logger } from "../utils/logger.js";
import { setSetting, getSettingValue } from "../bot/telegram-bot.js";

let twitterClient: TwitterApi | null = null;

function getTwitterClient(): TwitterApi {
  if (!twitterClient) {
    twitterClient = new TwitterApi(config.TWITTER_BEARER);
  }
  return twitterClient;
}

export async function crawlTwitterSources(): Promise<number> {
  if (!config.hasTwitter) {
    return 0;
  }

  const sources = await db.source.findMany({
    where: { type: "twitter", isActive: true },
  });

  if (sources.length === 0) return 0;

  let totalNew = 0;
  const client = getTwitterClient();

  for (const source of sources) {
    try {
      const count = await crawlOneAccount(client, source);
      totalNew += count;
    } catch (err: any) {
      if (err.code === 429) {
        logger.warn("twitter", `Rate limited. Skipping cycle.`);
        break; // Stop all Twitter crawling this cycle
      }
      logger.error("twitter", `Error crawling ${source.handle}: ${err.message}`);
    }
  }

  if (totalNew > 0) {
    logger.info("twitter", `Crawled ${totalNew} new tweets total`);
  }

  return totalNew;
}

async function crawlOneAccount(
  client: TwitterApi,
  source: { id: string; handle: string; name: string }
): Promise<number> {
  const handle = source.handle.replace("@", "");
  const settingKey = `tw_since_${handle}`;
  const sinceId = await getSettingValue(settingKey, "");

  // Get user ID from username
  const userIdKey = `tw_uid_${handle}`;
  let userId = await getSettingValue(userIdKey, "");

  if (!userId) {
    try {
      const user = await client.v2.userByUsername(handle);
      userId = user.data.id;
      await setSetting(userIdKey, userId);
    } catch (err: any) {
      logger.error("twitter", `Cannot find user @${handle}: ${err.message}`);
      return 0;
    }
  }

  // Lần đầu crawl: lấy tweet mới nhất làm mốc, KHÔNG lấy tin cũ
  if (!sinceId) {
    const firstFetch = await client.v2.userTimeline(userId, {
      max_results: 5,
      "tweet.fields": ["created_at"],
      exclude: ["retweets", "replies"],
    });
    const firstTweet = firstFetch.data?.data?.[0];
    if (firstTweet) {
      await setSetting(settingKey, firstTweet.id);
      logger.info("twitter", `@${handle}: initialized at tweet #${firstTweet.id} (skip old tweets)`);
    }
    return 0;
  }

  // Fetch tweets since last known
  const params: any = {
    max_results: 10,
    "tweet.fields": ["created_at", "author_id", "text"],
    exclude: ["retweets", "replies"],
    since_id: sinceId,
  };

  const tweets = await client.v2.userTimeline(userId, params);

  let newCount = 0;
  let newestId = sinceId;

  for (const tweet of tweets.data?.data || []) {
    if (!tweet.text || tweet.text.trim().length < 10) continue;

    // Track newest
    if (!newestId || BigInt(tweet.id) > BigInt(newestId)) {
      newestId = tweet.id;
    }

    try {
      await db.contentItem.create({
        data: {
          sourceId: source.id,
          externalId: tweet.id,
          originalText: tweet.text,
          authorName: `@${handle}`,
          sourceUrl: `https://x.com/${handle}/status/${tweet.id}`,
          status: "pending",
        },
      });
      newCount++;
    } catch (err: any) {
      if (err.code !== "P2002") {
        logger.error("twitter", `DB error: ${err.message}`);
      }
    }
  }

  if (newestId && newestId !== sinceId) {
    await setSetting(settingKey, newestId);
  }

  if (newCount > 0) {
    logger.info("twitter", `@${handle}: ${newCount} new tweets`);
  }

  return newCount;
}
