package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type appConfig struct {
	Days    int
	Retweet bool
	Twitter TwitterConfig
}

func loadConfig(filepath string) appConfig {
	var config appConfig
	_, err := toml.DecodeFile(filepath, &config)
	if err != nil {
		log.Fatal("config load error:", err)
	}
	return config
}

func saveConfig(filepath string, config appConfig, twitter *Twitter) {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	config.Twitter = twitter.config
	err := encoder.Encode(config)
	if err != nil {
		log.Fatal("failed to encode config:", err)
	}

	err = ioutil.WriteFile(filepath, buffer.Bytes(), os.ModePerm)
	if err != nil {
		log.Fatal("failed to store config file:", err)
	}
}

func main() {
	configFile := flag.String("config", "retrobot.toml", "config file")
	flag.Parse()

	config := loadConfig(*configFile)
	log.Printf("%#v", config)
	twitter := NewTwitter(config.Twitter)
	defer saveConfig(*configFile, config, twitter)

	csvfile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("can't open csv:", err)
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("csv read error:", err)
	}

	for i := len(records) - 1; i >= 0; i-- {
		tweet := records[i]
		ts, err := time.Parse("2006-01-02 15:04:05 -0700", tweet[3])
		if err != nil {
			log.Fatal("time parse error:", err)
		}
		if ts.After(time.Now().AddDate(0, 0, -config.Days)) {
			for {
				time.Sleep(1 * time.Second)
				if ts.Before(time.Now().AddDate(0, 0, -config.Days)) {
					if tweet[6] == "" {
						twitter.PostTweet(strings.Replace(tweet[5], "@", "", -1))
						log.Printf("%s", tweet[5])
					} else {
						if config.Retweet {
							twitter.Retweet(tweet[6])
							log.Printf("retweet: (%s) %s", tweet[6], tweet[5])
						} else {
							log.Printf("skip retweet: %s", tweet[6])
						}
					}
					break
				}
			}
		}
	}
}
