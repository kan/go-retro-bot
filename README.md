go-retro-bot
====

[mirakui/retrobot](https://github.com/mirakui/retrobot)のgolang実装

## Usage

```sh
go get https://github.com/kan/go-retro-bot.git
curl -LO https://raw.githubusercontent.com/kan/go-retro-bot/master/retrobot.toml.example
# edit retrobot.toml.example
mv retrobot.toml.example retrobot.toml
# if you don't have tweets.csv, see https://blog.twitter.com/2012/your-twitter-archive
go-retro-bot -c ./retrobot.toml tweets.csv
```

## TODO

- 現在時刻とretro_daysにあわせたツイートの投稿
 - reply_to_urlを追加するオプションに対応
- 単体でdaemonとして常駐可能に
- 気力があればリリースフローを整理

## Licence

[MIT](https://github.com/kan/go-retro-bot/blob/master/LICENSE)

## Author

[kan](https://github.com/kan)

