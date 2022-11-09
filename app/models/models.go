package models

type Wish struct {
	Id     int    `json:"id" gorm:"primaryKey"`
	Name   string `json:"Name"`
	Price  int    `json:"Price ($)"`
	Link   string `json:"Link"`
	UserId int
	User   User
}

type AirTableConnection struct {
	Id    int
	Token string
	Board string
	Path  string
}

type User struct {
	Id         int                `gorm:"primaryKey"`
	ChatId     int64              `gorm:"unique;not null"`
	AirTable   AirTableConnection `gorm:"embedded"`
	AirTableId int
}
