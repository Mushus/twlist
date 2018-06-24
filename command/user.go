package command

import (
	"bufio"
	"fmt"
	"os"

	tw "github.com/Mushus/go-twitter/twitter"
	"github.com/Mushus/twlist/storage"
	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/jinzhu/gorm"
	"github.com/pkg/browser"
)

// AuthenticationTwitter ツイッター認証をする
func AuthenticationTwitter(db *gorm.DB) error {
	config := oauth1.Config{
		ConsumerKey:    "q2GA9KHQ4AuhfZKyoKiZKMKMy",
		ConsumerSecret: "YLG2ZwlgWTnz2WHBQRfK29UXxnNp0TPUN6OGkvOAl7sE0hvdZN",
		CallbackURL:    "",
		Endpoint:       twitter.AuthorizeEndpoint,
	}
	requestToken, _, err := config.RequestToken()
	if err != nil {
		return fmt.Errorf("failed to get request token: %v", err)
	}
	authorizationURL, err := config.AuthorizationURL(requestToken)
	if err != nil {
		return fmt.Errorf("failed to get authorization url: %v", err)
	}
	err = browser.OpenURL(authorizationURL.String())
	if err != nil {
		fmt.Printf("Authorization URL: %v\n", authorizationURL)
	}
	fmt.Print("pin? ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	verifier := scanner.Text()
	accessToken, accessSecret, err := config.AccessToken(requestToken, "", verifier)
	if err != nil {
		return fmt.Errorf("failed to get request token: %v", err)
	}

	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := tw.NewClient(httpClient)
	me, _, err := client.Accounts.VerifyCredentials(&tw.AccountVerifyParams{})
	if err != nil {
		return fmt.Errorf("failed to get User information: %v", err)
	}

	var count int
	db.Model(&storage.MyUser{}).Count(&count)
	defaultUser := count == 0
	user := storage.MyUser{
		ScreenName: me.ScreenName,
		Default:    defaultUser,
		Authentication: storage.Authentication{
			ConsumerKey:    config.ConsumerKey,
			ConsumerSecret: config.ConsumerSecret,
			AccessToken:    accessToken,
			AccessSecret:   accessSecret,
		},
	}

	db.Create(&user)

	return nil
}
