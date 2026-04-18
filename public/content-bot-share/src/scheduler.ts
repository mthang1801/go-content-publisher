import { config } from "./config.js";
import { logger } from "./utils/logger.js";
import { crawlTelegramSources } from "./crawlers/telegram-crawler.js";
import { crawlTwitterSources } from "./crawlers/twitter-crawler.js";
import { processNextPending } from "./processor/deepseek.js";
import { publishNextReady, publishNextToTwitter } from "./publisher/telegram-publisher.js";

let crawlTimer: NodeJS.Timeout | null = null;
let processTimer: NodeJS.Timeout | null = null;
let publishTimer: NodeJS.Timeout | null = null;
let twitterTimer: NodeJS.Timeout | null = null;

async function crawlCycle() {
  try {
    await crawlTelegramSources();
  } catch (err: any) {
    logger.error("scheduler", `TG crawl error: ${err.message}`);
  }

  try {
    await crawlTwitterSources();
  } catch (err: any) {
    logger.error("scheduler", `Twitter crawl error: ${err.message}`);
  }
}

async function processCycle() {
  try {
    await processNextPending();
  } catch (err: any) {
    logger.error("scheduler", `Process error: ${err.message}`);
  }
}

async function publishCycle() {
  try {
    await publishNextReady();
  } catch (err: any) {
    logger.error("scheduler", `Publish error: ${err.message}`);
  }
}

async function twitterCycle() {
  try {
    await publishNextToTwitter();
  } catch (err: any) {
    logger.error("scheduler", `Twitter publish error: ${err.message}`);
  }
}

export function startScheduler() {
  logger.info("scheduler", `Starting scheduler:`);
  logger.info("scheduler", `  Crawl: every ${config.CRAWL_INTERVAL}s`);
  logger.info("scheduler", `  Process: every ${config.PROCESS_INTERVAL}s`);
  logger.info("scheduler", `  Publish: every ${config.PUBLISH_INTERVAL}s`);
  logger.info("scheduler", `  Twitter: every ${config.TWITTER_PUBLISH_INTERVAL}s`);

  // Run first crawl after 10s delay (let bot initialize)
  setTimeout(crawlCycle, 10_000);

  crawlTimer = setInterval(crawlCycle, config.CRAWL_INTERVAL * 1000);
  processTimer = setInterval(processCycle, config.PROCESS_INTERVAL * 1000);
  publishTimer = setInterval(publishCycle, config.PUBLISH_INTERVAL * 1000);
  twitterTimer = setInterval(twitterCycle, config.TWITTER_PUBLISH_INTERVAL * 1000);
}

export function stopScheduler() {
  if (crawlTimer) clearInterval(crawlTimer);
  if (processTimer) clearInterval(processTimer);
  if (publishTimer) clearInterval(publishTimer);
  if (twitterTimer) clearInterval(twitterTimer);
  crawlTimer = processTimer = publishTimer = twitterTimer = null;
  logger.info("scheduler", "Scheduler stopped");
}

// Manual trigger
export { crawlCycle as triggerCrawl };
