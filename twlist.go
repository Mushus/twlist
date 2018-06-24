package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Mushus/go-twitter/twitter"
	"github.com/Mushus/pigeonhole/command"
	"github.com/Mushus/pigeonhole/storage"
	"github.com/dghubble/oauth1"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	subcommand := ""
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
	}
	subsubcommand := ""
	if len(os.Args) > 2 {
		subsubcommand = os.Args[2]
	}

	if subcommand == "" {
		// usage
		return
	}

	filename, err := storage.GetDBPath()
	if err != nil {
		log.Fatalf("cannot get db path: %v", err)
	}

	db, err := gorm.Open("sqlite3", filename)
	storage.Migration(db)
	defer db.Close()

	// ユーザーアカウントの操作
	if subcommand == "user" {
		if subsubcommand == "add" {
			command.AuthenticationTwitter(db)
			return
		}
		// usage
		return
	}

	// ユーザー情報を用意する
	tc, err := CreateTwitterClient(db)
	if err != nil {
		log.Fatalf("failed to connect twitter: %v", err)
	}

	switch subcommand {
	case "collect":
		switch subsubcommand {
		case "friendship":
			err := command.CollectFriendship(db, tc)
			if err != nil {
				log.Fatalf("failed to collect friendship: %v", err)
			}
		}
	case "listgen":
		switch subsubcommand {
		case "friendship":
			err := command.ListgenFriendship(db, tc)
			if err != nil {
				log.Fatalf("failed to create friendship list: %v", err)
			}
		}
	}
}

// CreateTwitterClient twitterクライアントを作成する
func CreateTwitterClient(db *gorm.DB) (*twitter.Client, error) {
	me := storage.MyUser{}
	err := db.Order("default").Preload("Authentication").First(&me).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	auth := me.Authentication
	config := oauth1.NewConfig(auth.ConsumerKey, auth.ConsumerSecret)
	token := oauth1.NewToken(auth.AccessToken, auth.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	return client, nil
}
