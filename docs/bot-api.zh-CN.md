# 在本服使用机器人

本服不是官方 Telegram。机器人只能在这台 gramsrv 上创建，token 也只能打这台服务的 Bot API。官方 `api.telegram.org` 的 token 在这里无效，这里发出的 token 也不能拿去打官方接口。

方法、参数和错误码的完整清单见 [bot-api.en.md](./bot-api.en.md)（俄文对照 [bot-api.ru.md](./bot-api.ru.md)）。那份清单把 `sendMediaGroup` 列为未实现；`local` 分支已经实现相册发送。

## 这台机器的地址

| 项 | 值 |
|---|---|
| Bot API 根地址 | `https://bot.zhufu.de5.net` |
| 调用方法 | `https://bot.zhufu.de5.net/bot<TOKEN>/<METHOD>` |
| 下载文件 | `https://bot.zhufu.de5.net/file/bot<TOKEN>/<file_path>` |
| 监听 | `TELESRV_BOT_API_ADDR`，空值表示关闭网关 |
| BotFather 附带展示的地址 | `TELESRV_BOT_API_PUBLIC_URL`，空值则不展示 |

`<TOKEN>` 形如 `<bot_id>:<secret>`，由本服 `@BotFather` 发出。不要把 token 写进仓库、日志或截图。

确认网关和 token：

```bash
curl -sS "https://bot.zhufu.de5.net/bot<TOKEN>/getMe"
```

返回 `"ok": true` 且 `is_bot` 为 true 即可。`0.0.0.0` 只能作监听地址，不能当客户端要请求的主机。

## 创建机器人

1. 用已经连上这台服务的客户端打开 `@BotFather`。必须是本服里的 BotFather，不是官方 Telegram 里的那个。
2. 发送 `/newbot`，按提示填写显示名和以 `bot` 结尾的用户名。
3. 保存返回的 token。若配置了 `TELESRV_BOT_API_PUBLIC_URL`，同一条回复里会带上 API 地址。
4. 在客户端里打开这个机器人并发送 `/start`。没有先打开对话，机器人不能主动发私聊。

同一个 token 同时只能有一个程序调用 `getUpdates`。第二个会收到 HTTP 409，错误标记为 `CONFLICT`。改用 webhook 之后不要再开长轮询。

## 把程序指到本服

配置里填根地址，不要自己加上 `/bot` 或 token，除非该库的字段明确要求带到 `/bot`。

| 库 | 配置 |
|---|---|
| python-telegram-bot | `base_url` 为 `https://bot.zhufu.de5.net/bot`，`base_file_url` 为 `https://bot.zhufu.de5.net/file/bot` |
| aiogram 3 | `TelegramAPIServer.from_base("https://bot.zhufu.de5.net")` |
| grammY | `apiRoot: "https://bot.zhufu.de5.net/bot"` |
| QwenPaw | 频道里的 `base_url` 填 `https://bot.zhufu.de5.net`，`http_proxy` 留空 |

QwenPaw 会自己在根地址后面拼 `/bot` 和 `/file/bot`。`base_url` 留空时它去请求 `api.telegram.org`，本服 token 会被拒绝。它实际用到、并且本服已经实现的接口有 `getUpdates`、`sendMessage`、`editMessageText`、`sendChatAction`、`deleteMessage`、`sendPhoto`、`sendDocument` 和 `getFile`。文本可以用 `parse_mode` 的 `HTML`、`Markdown` 或 `MarkdownV2`。

```python
from telegram.ext import Application

app = (
    Application.builder()
    .token("<TOKEN>")
    .base_url("https://bot.zhufu.de5.net/bot")
    .base_file_url("https://bot.zhufu.de5.net/file/bot")
    .build()
)
```

仓库里可直接跑的例子：

| 路径 | 作用 |
|---|---|
| `cmd/bots/ptbecho` | python-telegram-bot 回声，含 ephemeral 命令 |
| `cmd/bots/aiogramecho` | aiogram 回声、回复键盘和内联按钮 |
| `cmd/bots/bedolagaformat` | HTML / Markdown、富文本菜单、Telegram Login |
| `cmd/bots/grammystore` | 跑在 HTTP Bot API 上的 grammY 业务机器人 |

## Webhook

长轮询是程序来拉本服。Webhook 是本服把更新 POST 到你登记的地址，方向相反。地址怎么区分、`getWebhookInfo` 怎么查，见 [configuration.zh-CN.md](./configuration.zh-CN.md) 的「Bot API webhook 故障排查」。
