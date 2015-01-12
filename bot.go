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

type Config struct {
	Days    int
	Twitter TwitterConfig
}

func loadConfig(filepath string) Config {
	var config Config
	_, err := toml.DecodeFile(filepath, &config)
	if err != nil {
		log.Fatal("config load error:", err)
	}
	return config
}

func saveConfig(filepath string, config Config, twitter *Twitter) {
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
	config_file := flag.String("config", "retrobot.toml", "config file")
	flag.Parse()

	config := loadConfig(*config_file)
	twitter := NewTwitter(config.Twitter)
	defer saveConfig(*config_file, config, twitter)

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
					twitter.PostTweet(strings.Replace(tweet[5], "@", "", -1))
					log.Printf("%s", tweet[5])
					break
				}
			}
		}
	}
}
