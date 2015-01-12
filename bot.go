package main

import (
	"flag"
)

func main() {
	config_file := flag.String("config", "retrobot.toml", "config file")
	flag.Parse()

	twitter := NewTwitter()
	twitter.LoadConfig(*config_file)
	defer twitter.SaveConfig(*config_file)
	status := flag.Arg(0)
	twitter.PostTweet(status)
}
