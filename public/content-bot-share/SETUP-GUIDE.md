# Content Bot - Auto Crawl & Publish

Bot tu dong crawl tin tu Telegram + Twitter, dung DeepSeek AI viet lai tieng Viet, dang len Telegram Channel + Twitter.

## Pipeline

```
Telegram Channels ─┐
                    ├─→ Ro Content ─→ DeepSeek AI ─→ Telegram Channel
Twitter Accounts  ─┤      (DB)       (viet lai)      Twitter VI
                    │                                  Twitter EN
Manual /add       ─┘
```

## Tinh Nang

- Crawl tu dong tu nhieu nguon Telegram + Twitter
- AI viet lai bang tieng Viet (DeepSeek)
- Tu dong dich sang tieng Anh cho Twitter EN
- Loc quang cao, spam, link affiliate
- Kiem chung thong tin truoc khi dang
- Loc trung lap 70% (bigram similarity)
- Gop tin tu nguon dang nhieu (VN Wallstreet)
- Quan ly qua Telegram bot (/add, /status, /sources...)
- Dang len Telegram Channel + 2 tai khoan Twitter

---

## Huong Dan Cai Dat

### Buoc 1: Cai Dat Moi Truong

**Yeu cau:**
- Node.js v20+ (https://nodejs.org)
- pnpm (package manager)

```bash
# Cai pnpm
npm install -g pnpm

# Clone/download project
cd content-bot
pnpm install
```

### Buoc 2: Tao Telegram Bot

1. Mo Telegram, tim **@BotFather**
2. Gui `/newbot` → dat ten → nhan **Bot Token**
3. Tao Channel moi (hoac dung channel co san)
4. Add bot vao Channel lam **Admin** (quyen Post Messages)
5. Lay **Channel ID**: forward 1 tin tu channel den **@userinfobot**
6. Lay **User ID** cua ban: gui tin den **@userinfobot**

### Buoc 3: Tao DeepSeek API Key

1. Vao https://platform.deepseek.com
2. Dang ky tai khoan
3. Vao API Keys → tao key moi
4. Nap tien (rat re, ~$1-2/thang)

### Buoc 4: Cau Hinh .env

```bash
cp .env.example .env
```

Mo file `.env` va dien thong tin:

```env
# BAT BUOC
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_TARGET_CHANNEL=-100xxxxxxxxxx
ADMIN_USER_IDS=your_user_id
DEEPSEEK_API_KEY=sk-xxxxxxxx
```

### Buoc 5: Khoi Tao Database

```bash
npx prisma generate
npx prisma db push
```

### Buoc 6: Chay Bot

```bash
# Development (co hot-reload)
pnpm dev

# Production voi PM2
npm install -g pm2 tsx
pm2 start src/index.ts --name content-bot --interpreter tsx
pm2 save
```

---

## Cau Hinh Nang Cao (Tuy Chon)

### Telegram Crawler (doc tin tu channel)

1. Vao https://my.telegram.org/apps
2. Tao app → lay **API ID** va **API Hash**
3. Dien vao `.env`:
```env
TELEGRAM_API_ID=12345678
TELEGRAM_API_HASH=abcdef1234567890
```
4. Tao session:
```bash
pnpm auth:telegram
```
5. Copy session string vao `.env` dong `TELEGRAM_SESSION=`
6. Restart bot

### Twitter Crawl (doc tweets)

1. Vao https://developer.x.com → tao app
2. Lay **Bearer Token** tu Keys and Tokens
3. Dien vao `.env`:
```env
TWITTER_BEARER_TOKEN=AAAA...
```

### Twitter Publish (dang tweet)

Can 2 tai khoan Twitter Developer (1 VI, 1 EN):

1. Tao app tren https://developer.x.com
2. App permissions: **Read and Write**
3. User authentication: **Web App, Automated App or Bot**
4. Callback URL: `https://localhost`
5. Generate **OAuth 1.0a** keys:
   - API Key (Consumer Key)
   - API Key Secret (Consumer Secret)
   - Access Token
   - Access Token Secret

6. Dien vao `.env`:
```env
# Tai khoan tieng Viet
TWITTER_VI_API_KEY=xxx
TWITTER_VI_API_SECRET=xxx
TWITTER_VI_ACCESS_TOKEN=xxx
TWITTER_VI_ACCESS_SECRET=xxx

# Tai khoan tieng Anh
TWITTER_EN_API_KEY=xxx
TWITTER_EN_API_SECRET=xxx
TWITTER_EN_ACCESS_TOKEN=xxx
TWITTER_EN_ACCESS_SECRET=xxx
```

---

## Su Dung Bot

### Lenh Telegram Bot

| Lenh | Mo ta |
|------|-------|
| `/start` | Hien menu |
| `/add <text>` | Them noi dung vao Ro Content |
| `/status` | Xem tong quan |
| `/queue` | Xem hang doi |
| `/recent` | Bai da dang gan day |
| `/sources` | Danh sach nguon |
| `/addsource <type> <handle> <name>` | Them nguon moi |
| `/removesource <handle>` | Xoa nguon |
| `/crawlnow` | Crawl ngay lap tuc |
| `/pause` | Tam dung auto-publish |
| `/resume` | Bat lai auto-publish |
| `/logs` | Xem log gan day |
| `/retry` | Thu lai cac item loi |

### Them Nguon

```
/addsource telegram @channelname Ten Hien Thi
/addsource twitter @username Ten Hien Thi
```

### Gui Tin Thu Cong

Gui bat ky tin nhan nao (khong co /) den bot → tu dong them vao Ro Content.

Hoac dung:
```
/add Noi dung tin tuc can dang...
```

---

## Deploy len VPS

### Windows Server

```powershell
# 1. Cai Node.js tu https://nodejs.org
# 2. Copy project len VPS
# 3. Cai dat
cd C:\content-bot
npm install -g pnpm pm2 tsx
pnpm install
npx prisma generate
npx prisma db push

# 4. Chay
pm2 start src\index.ts --name content-bot --interpreter tsx
pm2 save
```

### Linux (Ubuntu/Debian)

```bash
# 1. Cai Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# 2. Cai tools
npm install -g pnpm pm2 tsx

# 3. Setup project
cd /opt/content-bot
pnpm install
npx prisma generate
npx prisma db push

# 4. Chay
pm2 start src/index.ts --name content-bot --interpreter tsx
pm2 save
pm2 startup
```

---

## Prompt Goc Cho Claude Code

Neu ban muon dung Claude Code de tu tao bot nay tu dau, day la prompt:

```
Tao mot Content Bot voi cac module sau:

1. TELEGRAM BOT (node-telegram-bot-api):
   - Nhan lenh tu admin qua Telegram
   - /add <text> - them noi dung vao "Ro Content" (SQLite DB)
   - /status, /queue, /sources, /addsource, /removesource
   - Gui tin nhan thuong (khong co /) → tu dong them vao Ro Content
   - Chi cho phep admin (ADMIN_USER_IDS)

2. TELEGRAM CRAWLER (GramJS/telegram package):
   - Crawl tin moi tu cac channel Telegram da chi dinh
   - Realtime listener + polling backup moi 60s
   - Gop tin tu nguon dang nhieu (buffer 2-5 phut) thanh 1 bai tong hop
   - Chi lay tin moi (khong crawl tin cu)
   - Luu vao Ro Content voi status "pending"

3. TWITTER CRAWLER (twitter-api-v2):
   - Dung Bearer token doc tweets tu cac account chi dinh
   - Crawl moi 60s, chi lay tweets moi (theo since_id)
   - Luu vao Ro Content

4. DEEPSEEK PROCESSOR (OpenAI SDK, baseURL deepseek):
   - Lay item "pending" tu Ro Content
   - Gui qua DeepSeek API de:
     a) Viet lai thanh bai tin tuc tieng Viet (100-300 tu)
     b) Dich sang tieng Anh
     c) Tao phien ban ngan cho Twitter (max 4 dong, 250 ky tu) ca VI va EN
     d) Kiem chung thong tin
     e) Loc quang cao/spam/link → tu dong reject
   - Tra ve JSON: rewrittenText, rewrittenTextEn, tweetVI, tweetEN, factCheckNote, shouldPublish
   - Duplicate detection: bigram similarity 70%, so sanh ca originalText va rewrittenText
   - Goc nhin khach quan cho thi truong tai chinh toan cau (khong thien ve crypto)

5. PUBLISHER:
   - Telegram: dang rewrittenText len channel (Markdown format)
   - Fallback plain text neu Markdown loi
   - Twitter VI: dang tweetVI len tai khoan tieng Viet (OAuth 1.0a)
   - Twitter EN: dang tweetEN len tai khoan tieng Anh
   - Twitter publish tach rieng, chay cham hon (moi 20s)
   - Loai bo tat ca link, ten nguon, @handle truoc khi dang

6. SCHEDULER:
   - Crawl: moi 60s
   - Process (DeepSeek): moi 15s
   - Publish (Telegram): moi 5s
   - Publish (Twitter): moi 20s

TECH STACK:
- TypeScript, tsx (runtime)
- SQLite + Prisma ORM
- node-telegram-bot-api (bot)
- telegram/GramJS (crawler)
- twitter-api-v2 (crawl + publish)
- openai SDK (DeepSeek compatible)
- dotenv cho config
- PM2 cho production

YEU CAU:
- Tat ca config qua .env (khong hardcode credentials)
- Tao .env.example day du
- Bot khong crash khi loi API (try/catch tat ca)
- Log day du voi timestamp + module name
- Tu dong skip tin cu khi crawl lan dau
- Strip tat ca URL, ten nguon, @handle khoi bai dang
- Content dau vao nhieu ngon ngu, dau ra chi tieng Viet
- Ngon ngu giao tiep bot: tieng Viet
```

---

## Cau Truc Thu Muc

```
content-bot/
├── .env.example          # Template config
├── .env                  # Config thuc (KHONG SHARE)
├── .gitignore
├── package.json
├── tsconfig.json
├── prisma/
│   └── schema.prisma     # Database schema
├── scripts/
│   └── telegram-auth.ts  # Script tao Telegram session
├── data/
│   └── bot.db            # SQLite database (tu tao)
└── src/
    ├── index.ts           # Entry point
    ├── config.ts          # Load env vars
    ├── db.ts              # Prisma client
    ├── scheduler.ts       # Task scheduler
    ├── bot/
    │   └── telegram-bot.ts
    ├── crawlers/
    │   ├── telegram-crawler.ts
    │   └── twitter-crawler.ts
    ├── processor/
    │   └── deepseek.ts
    ├── publisher/
    │   ├── telegram-publisher.ts
    │   └── twitter-publisher.ts
    └── utils/
        └── logger.ts
```

---

## Luu Y

- **KHONG BAO GIO** share file `.env` (chua API keys)
- DeepSeek API rat re (~$0.14/1M tokens input, $0.28/1M output)
- Twitter Free tier: 1,500 tweets/thang/app
- Bot tu dong skip tin trung lap 70%+
- Neu bot crash, PM2 tu dong restart
- Kiem tra logs: `pm2 logs content-bot`

---

Built with Claude Code | Powered by DeepSeek AI
