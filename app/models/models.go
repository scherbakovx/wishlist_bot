package models

import "fmt"

type Wish struct {
	Name  string `json:"Name"`
	Price int    `json:"Price ($)"`
	Link  string `json:"Link"`
}

type LocalWish struct {
	Wish
	Id     int `json:"id" gorm:"primaryKey"`
	UserId int
	User   User
}

// String returns string-representation of Wish-object
func (w Wish) String() string {
	return fmt.Sprintf("Name: %s\nPrice: %d$\nLink: %s", w.Name, w.Price, w.Link)
}

type AirTableConnection struct {
	Id    int
	Token string
	Board string
	Path  string
}

type UserStatus int

const (
	Writer UserStatus = iota
	Reader
)

type User struct {
	Id     int   `gorm:"primaryKey"`
	ChatId int64 `gorm:"unique;not null"`

	AirTable   AirTableConnection `gorm:"embedded"`
	AirTableId int

	Status        UserStatus `gorm:"default:0"`
	ReadingUserId int
}
