import { TwitterApi } from "twitter-api-v2";
import { config } from "../config.js";
import { logger } from "../utils/logger.js";

let twitterViClient: TwitterApi | null = null;
let twitterEnClient: TwitterApi | null = null;

function getViClient(): TwitterApi | null {
  if (!config.hasTwitterPublishVI) return null;
  if (!twitterViClient) {
    twitterViClient = new TwitterApi({
      appKey: config.TWITTER_VI_API_KEY,
      appSecret: config.TWITTER_VI_API_SECRET,
      accessToken: config.TWITTER_VI_ACCESS_TOKEN,
      accessSecret: config.TWITTER_VI_ACCESS_SECRET,
    });
  }
  return twitterViClient;
}

function getEnClient(): TwitterApi | null {
  if (!config.hasTwitterPublishEN) return null;
  if (!twitterEnClient) {
    twitterEnClient = new TwitterApi({
      appKey: config.TWITTER_EN_API_KEY,
      appSecret: config.TWITTER_EN_API_SECRET,
      accessToken: config.TWITTER_EN_ACCESS_TOKEN,
      accessSecret: config.TWITTER_EN_ACCESS_SECRET,
    });
  }
  return twitterEnClient;
}

// Twitter max 280 chars - cắt text thông minh
function truncateForTweet(text: string, maxLen: number = 275): string {
  // Remove Markdown formatting (*bold*, _italic_)
  let clean = text.replace(/\*([^*]+)\*/g, "$1");
  clean = clean.replace(/_([^_]+)_/g, "$1");
  // Remove multiple newlines
  clean = clean.replace(/\n{2,}/g, "\n\n");
  clean = clean.trim();

  if (clean.length <= maxLen) return clean;

  // Cắt tại câu cuối cùng trong limit
  const truncated = clean.slice(0, maxLen);
  const lastSentence = truncated.lastIndexOf(".");
  const lastNewline = truncated.lastIndexOf("\n");
  const cutPoint = Math.max(lastSentence, lastNewline);

  if (cutPoint > maxLen * 0.5) {
    return truncated.slice(0, cutPoint + 1).trim() + "…";
  }
  return truncated.trim() + "…";
}

export async function publishToTwitterVI(text: string): Promise<string | null> {
  const client = getViClient();
  if (!client) return null;

  try {
    const tweetText = truncateForTweet(text);
    const result = await client.v2.tweet(tweetText);
    logger.info("twitter-vi", `Tweeted VI: ${result.data.id}`);
    return result.data.id;
  } catch (err: any) {
    const detail = err.data ? JSON.stringify(err.data) : err.message;
    logger.error("twitter-vi", `Tweet VI failed: ${detail}`);
    return null;
  }
}

export async function publishToTwitterEN(text: string): Promise<string | null> {
  const client = getEnClient();
  if (!client) return null;

  try {
    const tweetText = truncateForTweet(text);
    const result = await client.v2.tweet(tweetText);
    logger.info("twitter-en", `Tweeted EN: ${result.data.id}`);
    return result.data.id;
  } catch (err: any) {
    const detail = err.data ? JSON.stringify(err.data) : err.message;
    logger.error("twitter-en", `Tweet EN failed: ${detail}`);
    return null;
  }
}
