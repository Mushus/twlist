package command

import (
	"fmt"

	"github.com/Mushus/go-twitter/twitter"
	"github.com/Mushus/twlist/storage"
	"github.com/jinzhu/gorm"
)

type TwitterUser struct {
	ID             int64
	FollowersCount int
	FriendsCount   int
}

// GetMe 自分を取得する
func GetMe(db *gorm.DB, tc *twitter.Client) (TwitterUser, error) {
	me, _, err := tc.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{})
	if err != nil {
		return TwitterUser{}, fmt.Errorf("failed to user info: %v", err)
	}
	tu := TwitterUser{
		ID:             me.ID,
		FollowersCount: me.FollowersCount,
		FriendsCount:   me.FriendsCount,
	}
	err = db.Save(&storage.User{
		ID:         me.ID,
		ScreenName: me.ScreenName,
	}).Error
	if err != nil {
		return tu, fmt.Errorf("failed to save my user: %v", err)
	}
	return tu, err
}
