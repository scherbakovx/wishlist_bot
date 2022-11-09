package db

import (
	"fmt"
	"log"
	"os"

	"github.com/scherbakovx/wishlist_bot/app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	pg_username := os.Getenv("POSTGRES_USER")
	pg_password := os.Getenv("POSTGRES_PASSWORD")
	pg_db := os.Getenv("POSTGRES_DB")

	dbURL := fmt.Sprintf("postgres://%s:%s@db:5432/%s", pg_username, pg_password, pg_db)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.Wish{})
	db.AutoMigrate(&models.AirTableConnection{})
	db.AutoMigrate(&models.User{})

	return db
}
