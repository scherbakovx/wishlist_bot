package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/wishes", Wishes)
	e.Logger.Fatal(e.Start(":80"))
}

// e.GET("/wishes", Wishes)
func Wishes(c echo.Context) error {
	// var database *gorm.DB = db.Init()
	// user_tg_id := c.QueryParam("user_tg_id")

	// var wishes []models.LocalWish
	// result := database.Clauses(clause.OnConflict{DoNothing: true}).Model(&models.LocalWish{}).Joins("JOIN users ON local_wishes.user_id = users.id").Where("users.chat_id = ?", user_tg_id).Find(&wishes)
	// if result.Error != nil {
	// 	if result.Error.Error() == "record not found" {
	// 		return c.String(http.StatusNotFound, "404 not found")
	// 	} else {
	// 		return c.String(http.StatusBadRequest, result.Error.Error())
	// 	}
	// } else {
	// 	return c.String(http.StatusOK, fmt.Sprint(wishes[0]))
	// }
	return c.String(http.StatusOK, "hello!")
}
