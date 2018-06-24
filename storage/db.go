package storage

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/jinzhu/gorm"
)

// GetDBPath 設定ファイルのパスを取得します
func GetDBPath() (configPath string, err error) {
	dir := os.Getenv("Home")
	if dir == "" {
		if runtime.GOOS == "windows" {
			dir = os.Getenv("APPDATA")
			if dir == "" {
				dir = os.Getenv("USERPROFILE")
				if dir == "" {
					return "", errors.New("cannot find user directory")
				}
			}
			dir = filepath.Join(dir, "twlist")
		} else {
			usr, err := user.Current()
			if err != nil {
				return "", fmt.Errorf("cannot find user directory: %v", err)
			}
			dir = filepath.Join(usr.HomeDir, ".twlist")
		}
	} else {
		dir = filepath.Join(dir, ".twlist")
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %v", err)
	}

	configPath = filepath.Join(dir, "data.db")
	return configPath, nil
}

// Migration db に対してマイグレーションを行う
func Migration(db *gorm.DB) {
	//db.LogMode(true)
	db.Exec("PRAGMA foreign_keys = ON")
	err := db.AutoMigrate(&Authentication{}, &MyUser{}, &User{}, &Friendship{}).Error
	if err != nil {
		fmt.Println(err)
	}
}

// MyUser ユーザー情報
type MyUser struct {
	ID             int64 `gorm:"primary_key;AUTO_INCREMENT"`
	Default        bool
	ScreenName     string
	Authentication Authentication
}

// Authentication 認証情報
type Authentication struct {
	ID             int64 `gorm:"primary_key;AUTO_INCREMENT"`
	MyUserID       int64 `gorm:"index:idx_my_user" sql:"type:integer REFERENCES my_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT"`
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

type User struct {
	ID         int64 `gorm:"primary_key"`
	ScreenName string
}

type Friendship struct {
	From int64 `gorm:"primary_key;index:idx_from_user" sql:"type:integer REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE"`
	To   int64 `gorm:"primary_key;index:idx_TO_user" sql:"type:integer REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE"`
}
