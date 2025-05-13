package models

import (
	"backend/utils/token"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null;" json:"password"`
}

// Save User オブジェクトをデータベースに保存する
func (u *User) Save() (*User, error) {
	err := DB.Create(u).Error
	if err != nil {
		return nil, err
	}
	return u, nil
}

// BeforeSave Userオブジェクトが保存される前に実行する
func (u *User) BeforeSave(*gorm.DB) error {
	// パスワードをハッシュ化する
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)

	// ユーザーネームを小文字に変換する
	u.Username = strings.ToLower(u.Username)

	return nil
}

// PrepareOutput パスワードを空文字にして出力準備をする
func (u *User) PrepareOutput() *User {
	u.Password = ""
	return u
}

func GenerateToken(username string, password string) (string, error) {
	var user User

	err := DB.Where("username = ?", username).First(&user).Error

	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return "", err
	}

	token, err := token.GenerateToken(user.ID)

	if err != nil {
		return "", err
	}

	return token, nil
}
