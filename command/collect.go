package command

import (
	"fmt"

	"github.com/Mushus/go-twitter/twitter"
	"github.com/Mushus/twlist/storage"
	"github.com/jinzhu/gorm"
	"gopkg.in/cheggaaa/pb.v2"
)

func CollectFriendship(db *gorm.DB, tc *twitter.Client) error {
	me, err := GetMe(db, tc)
	if err != nil {
		return fmt.Errorf("failed to get ID: %v", err)
	}

	fmt.Println("update following...")
	bar := pb.StartNew(int(me.FollowersCount))
	for cursor := int64(-1); cursor != 0; {
		follower, _, err := tc.Followers.IDs(&twitter.FollowerIDParams{
			UserID: me.ID,
			Count:  2048,
			Cursor: cursor,
		})
		if err != nil {
			return fmt.Errorf("failed to get follower: %v", err)
		}
		for _, id := range follower.IDs {
			err := db.Save(&storage.User{
				ID: id,
			}).Error
			if err != nil {
				return fmt.Errorf("failed save user: %v", err)
			}

			err = db.Save(&storage.Friendship{
				From: id,
				To:   me.ID,
			}).Error
			if err != nil {
				return fmt.Errorf("failed save user friendship: %v", err)
			}

			bar.Increment()
		}
		cursor = follower.NextCursor
	}
	bar.Finish()

	fmt.Println("update friends...")
	bar = pb.StartNew(int(me.FriendsCount))
	for cursor := int64(-1); cursor != 0; {
		follower, _, err := tc.Friends.IDs(&twitter.FriendIDParams{
			UserID: me.ID,
			Count:  2048,
			Cursor: cursor,
		})
		if err != nil {
			return fmt.Errorf("failed to get follower: %v", err)
		}
		for _, id := range follower.IDs {
			err := db.Save(&storage.User{
				ID: id,
			}).Error
			if err != nil {
				return fmt.Errorf("failed save user: %v", err)
			}

			err = db.Save(&storage.Friendship{
				From: me.ID,
				To:   id,
			}).Error
			if err != nil {
				return fmt.Errorf("failed save user friendship: %v", err)
			}

			bar.Increment()
		}
		cursor = follower.NextCursor
	}
	bar.Finish()
	fmt.Println("finish!")
	//tc.Friends.List()
	return nil
}
