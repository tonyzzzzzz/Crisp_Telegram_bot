# Crisp Telegram Bot

A telegram bot built with golang to help integrate Crisp into Telegram.

Currently Supports:
- Forward user messages from crisp to admins on telegram.
- Reply user messages directly on telegram.

Will Support:
- Integration with Slack
- Detailed visitor info

## Getting Started

1. Get your crisp API credentials from [Crisp API token generator](https://go.crisp.chat/account/token/)
1. Create a bot with [BotFather](https://t.me/botfather), save the token for later use.
1. Build & Run.

## Installing & Deployment

### Use prebuilt binary
Download from [release page]().

### Built on your own
`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build`

Replace GOOS GOARCH with your server architecture.

## License

This project is licensed under the MIT License.

