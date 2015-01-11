package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/garyburd/go-oauth/oauth"
)

type Config struct {
	Days           int
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func openUrl(url string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	}

	fmt.Printf("open %s\n", url)
}

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
	TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
}

func clientAuth(requestToken *oauth.Credentials) (*oauth.Credentials, error) {
	url := oauthClient.AuthorizationURL(requestToken, nil)

	openUrl(url)
	fmt.Println("enter PIN.")
	stdin := bufio.NewReader(os.Stdin)
	b, err := stdin.ReadBytes('\n')
	if err != nil {
		log.Fatal("canceled")
	}

	if b[len(b)-2] == '\n' {
		b = b[0 : len(b)-2]
	} else {
		b = b[0 : len(b)-1]
	}
	accessToken, _, err := oauthClient.RequestToken(http.DefaultClient, requestToken, string(b))
	if err != nil {
		log.Fatal("canceled")
	}
	return accessToken, nil
}

func getAccessToken(config Config) (*oauth.Credentials, error) {
	oauthClient.Credentials.Token = config.ConsumerKey
	oauthClient.Credentials.Secret = config.ConsumerSecret

	var token *oauth.Credentials
	if config.AccessToken != "" && config.AccessSecret != "" {
		token = &oauth.Credentials{config.AccessToken, config.AccessSecret}
	} else {
		requestToken, err := oauthClient.RequestTemporaryCredentials(http.DefaultClient, "", nil)
		if err != nil {
			log.Print("failed to request temporary credentials:", err)
			return nil, err
		}
		token, err = clientAuth(requestToken)
		if err != nil {
			log.Print("failed to request temporary credentials:", err)
			return nil, err
		}
	}

	return token, nil
}

func postTweet(token *oauth.Credentials, status string) error {
	param := make(url.Values)
	param.Set("status", status)
	apiURL := "https://api.twitter.com/1.1/statuses/update.json"
	oauthClient.SignParam(token, "POST", apiURL, param)
	res, err := http.PostForm(apiURL, url.Values(param))
	if err != nil {
		log.Println("failed to post tweet:", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("ailed to post tweet:", err)
		return err
	}

	return nil
}

func main() {
	config_file := flag.String("config", "retrobot.toml", "config file")
	flag.Parse()

	var config Config
	_, err := toml.DecodeFile(*config_file, &config)
	if err != nil {
		log.Fatal("config load error:", err)
	}

	token, err := getAccessToken(config)
	if err != nil {
		log.Fatal("failed to get access token:", err)
	}
	if config.AccessToken == "" || config.AccessSecret == "" {
		config.AccessToken = token.Token
		config.AccessSecret = token.Secret
		var buffer bytes.Buffer
		encoder := toml.NewEncoder(&buffer)
		err := encoder.Encode(config)
		if err != nil {
			log.Fatal("failed to encode config:", err)
		}

		err = ioutil.WriteFile(*config_file, buffer.Bytes(), os.ModePerm)
		if err != nil {
			log.Fatal("failed to store config file:", err)
		}
	}
	status := flag.Arg(0)
	postTweet(token, status)
}
