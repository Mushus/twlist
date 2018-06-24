package command

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Mushus/go-twitter/twitter"
	"github.com/jinzhu/gorm"
	"github.com/manifoldco/promptui"
	pb "gopkg.in/cheggaaa/pb.v2"
)

// ListgenFriendship リストを生成する
func ListgenFriendship(db *gorm.DB, tc *twitter.Client) error {
	fmt.Println("To perform this command, the following operations are needed.")
	fmt.Printf(" - %v collect friendship\n", os.Args[0])
	prompt := promptui.Select{
		Label: "Start?",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("falied to get user prompt result: %v", err)
	}

	if result != "Yes" {
		return nil
	}

	me, err := GetMe(db, tc)
	if err != nil {
		return fmt.Errorf("failed to get ID: %v", err)
	}

	lists, _, err := tc.Lists.List(&twitter.ListsListParams{
		UserID: me.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to get User ID")
	}
	var (
		funsListID      int64
		followingListID int64
	)
	for _, list := range lists {
		if list.Slug == "funs" {
			funsListID = list.ID
		}
		if list.Slug == "nonfollower" {
			followingListID = list.ID
		}
	}
	if funsListID == 0 {
		list, _, err := tc.Lists.Create("funs", &twitter.ListsCreateParams{
			Name:        "funs",
			Mode:        "private",
			Description: "Auto generated",
		})
		if err != nil {
			return fmt.Errorf("failed to crate funs list: %v", err)
		}
		funsListID = list.ID
	}
	if followingListID == 0 {
		list, _, err := tc.Lists.Create("nonfollower", &twitter.ListsCreateParams{
			Name:        "nonfollower",
			Mode:        "private",
			Description: "Auto generated",
		})
		if err != nil {
			return fmt.Errorf("failed to create following list: %v", err)
		}
		followingListID = list.ID
	}

	{ // ファンの人
		userIDs := []int64{}
		rows, err := db.
			Table("friendships as f1").
			Select("f1.`from`").
			Joins("left join friendships as f2 on f1.`from` = f2.`to` and f2.`from` = f1.`to`").
			Where("f1.`to` = ? and f2.`from` is null", me.ID).
			Rows()
		if err != nil {
			return fmt.Errorf("failed to get funs ids: %v", err)
		}
		for rows.Next() {
			var id int64
			err := rows.Scan(&id)
			if err != nil {
				return fmt.Errorf("failed to scan user id: %v", err)
			}
			userIDs = append(userIDs, id)
		}
		err = updateListUser(tc, userIDs, funsListID)
		if err != nil {
			return fmt.Errorf("failed to update funs list: %v", err)
		}
	}

	{ // 片道フォロー
		userIDs := []int64{}
		rows, err := db.
			Table("friendships as f1").
			Select("f1.`to`").
			Joins("left join friendships as f2 on f1.`from` = f2.`to` and f2.`from` = f1.`to`").
			Where("f1.`from` = ? and f2.`from` is null", me.ID).
			Rows()
		if err != nil {
			return fmt.Errorf("failed to get funs ids: %v", err)
		}
		for rows.Next() {
			var id int64
			err := rows.Scan(&id)
			if err != nil {
				return fmt.Errorf("failed to scan user id: %v", err)
			}
			userIDs = append(userIDs, id)
		}

		err = updateListUser(tc, userIDs, followingListID)
		if err != nil {
			return fmt.Errorf("failed to update funs list: %v", err)
		}
	}
	fmt.Println("finish!")
	return nil
}

func updateListUser(tc *twitter.Client, userIDs []int64, listID int64) error {
	listedIDsMap := map[int64]struct{}{}
	cursor := int64(-1)
	for cursor != 0 {
		listUsers, _, err := tc.Lists.Members(&twitter.ListsMembersParams{
			ListID: listID,
			Count:  5000,
		})
		if err != nil {
			return fmt.Errorf("failed to get list members: %v", err)
		}
		for _, user := range listUsers.Users {
			listedIDsMap[user.ID] = struct{}{}
		}
		cursor = listUsers.NextCursor
	}

	addedIDs := []string{}
	for _, id := range userIDs {
		_, ok := listedIDsMap[id]
		if ok {
			delete(listedIDsMap, id)
		} else {
			addedIDs = append(addedIDs, strconv.FormatInt(id, 10))
		}
	}
	deletedIDs := []string{}
	for id := range listedIDsMap {
		deletedIDs = append(deletedIDs, strconv.FormatInt(id, 10))
	}

	{
		bar := pb.StartNew(len(deletedIDs))
		counter := 0
		for counter < len(deletedIDs) {
			splitedIDs := []string{}
			i := 0
			for ; i < 100 && counter < len(deletedIDs); i++ {
				splitedIDs = append(splitedIDs, deletedIDs[counter])
				counter++
			}
			_, err := tc.Lists.MembersDestroyAll(&twitter.ListsMembersDestroyAllParams{
				ListID: listID,
				UserID: strings.Join(splitedIDs, ","),
			})
			if err != nil {
				return fmt.Errorf("failed to remove user from list: %v", err)
			}
			bar.Add(i)
		}
		bar.Finish()
	}

	{
		bar := pb.StartNew(len(addedIDs))
		counter := 0
		for counter < len(addedIDs) {
			splitedIDs := []string{}
			i := 0
			for ; i < 100 && counter < len(addedIDs); i++ {
				splitedIDs = append(splitedIDs, addedIDs[counter])
				counter++
			}
			_, err := tc.Lists.MembersCreateAll(&twitter.ListsMembersCreateAllParams{
				ListID: listID,
				UserID: strings.Join(splitedIDs, ","),
			})
			if err != nil {
				return fmt.Errorf("failed to add user from list: %v", err)
			}
			bar.Add(i)
		}
		bar.Finish()
	}

	return nil
}
