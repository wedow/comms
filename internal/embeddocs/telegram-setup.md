# Telegram Bot Setup

## 1. Create a Bot

1. Send `/newbot` to [@BotFather](https://t.me/BotFather) on Telegram.
2. Choose a display name (e.g. "My Comms Bot").
3. Choose a username ending in `bot` (e.g. `my_comms_bot`).
4. BotFather replies with an HTTP API token like `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`.

## 2. Configure comms

Add the token to `.comms/config.toml`:

```toml
[telegram]
token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
```

Or set the environment variable:

```sh
export COMMS_TELEGRAM_TOKEN="123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
```

The environment variable takes precedence over the config file.

## 3. Add Bot to a Group

1. Open the target Telegram group.
2. Add the bot by its username.
3. Send a message in the group so the bot can see the chat ID.

## 4. Get the Chat ID

Query the Telegram API to find your chat ID:

```sh
curl -s "https://api.telegram.org/bot<TOKEN>/getUpdates" | jq '.result[].message.chat.id'
```

Use this chat ID as the channel identifier when sending messages through comms.

## 5. Bot Privacy Mode

By default, bots only see messages that mention them or are commands. To receive all group messages:

1. Send `/setprivacy` to @BotFather.
2. Select your bot.
3. Choose "Disable".
