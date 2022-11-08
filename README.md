# Wishlist Bot
Bot to connect Tg bot-chat with Wishlist-table in AirTable // bot database DB

### TODO

- ~~Connect AirTable API~~
- ~~If link was shared to bot, then send it to board in AirTable~~
- ~~Connect something to read OpenGraph tags from links~~
- Connect PostgreSQL, so different users could create their wishlists
- Add feature, so friend of user could watch through user's wishes
- Add AirTable authentication, so everyone could create board there right from the bot

### Setup

- Create `.env` file in root directory with `TG_BOT_TOKEN` and `AIRTABLE_TOKEN` variables
- Run with:
    1. `go build app/main.go` -> `./main`
    2. `docker-compose up`