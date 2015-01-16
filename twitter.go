package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/garyburd/go-oauth/oauth"
)

// Twitter is simple twitter client
type Twitter struct {
	config      TwitterConfig
	oauthClient oauth.Client
	token       *oauth.Credentials
}

// TwitterConfig is config for twitter
type TwitterConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func openURL(url string) {
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

// NewTwitter is constractor for Twitter
func NewTwitter(config TwitterConfig) *Twitter {
	twitter := &Twitter{
		config: config,
		oauthClient: oauth.Client{
			TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
			ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
			TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		},
	}
	return twitter
}

func (t *Twitter) clientAuth(requestToken *oauth.Credentials) (*oauth.Credentials, error) {
	url := t.oauthClient.AuthorizationURL(requestToken, nil)

	openURL(url)
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
	accessToken, _, err := t.oauthClient.RequestToken(http.DefaultClient, requestToken, string(b))
	if err != nil {
		log.Fatal("canceled")
	}
	return accessToken, nil
}

func (t *Twitter) getAccessToken() error {
	t.oauthClient.Credentials.Token = t.config.ConsumerKey
	t.oauthClient.Credentials.Secret = t.config.ConsumerSecret

	if t.config.AccessToken != "" && t.config.AccessSecret != "" {
		t.token = &oauth.Credentials{t.config.AccessToken, t.config.AccessSecret}
	} else {
		requestToken, err := t.oauthClient.RequestTemporaryCredentials(http.DefaultClient, "", nil)
		if err != nil {
			log.Print("failed to request temporary credentials:", err)
			return err
		}
		token, err := t.clientAuth(requestToken)
		if err != nil {
			log.Print("failed to request temporary credentials:", err)
			return err
		}
		t.token = token
		t.config.AccessToken = token.Token
		t.config.AccessSecret = token.Secret
	}

	return nil
}

func (t *Twitter) post(apiURL string, param url.Values) error {
	if t.token == nil {
		err := t.getAccessToken()
		if err != nil {
			return err
		}
	}

	t.oauthClient.SignParam(t.token, "POST", apiURL, param)
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

// PostTweet post tweet to twitter API
func (t *Twitter) PostTweet(status string) error {
	param := make(url.Values)
	param.Set("status", status)
	apiURL := "https://api.twitter.com/1.1/statuses/update.json"
	return t.post(apiURL, param)
}

// Retweet to twitter API
func (t *Twitter) Retweet(statusID string) error {
	param := make(url.Values)
	apiURL := "https://api.twitter.com/1.1/statuses/retweet/" + statusID + ".json"
	return t.post(apiURL, param)
}
