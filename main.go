package main

import (
	"awesomeProject1/scrapper"
	"github.com/labstack/echo/v4"
	"os"
)

const fileName string = "jobs.csv"

// Handler
func handleHome(c echo.Context) error {
	//return c.String(http.StatusOK, "Hello, World!")
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := c.FormValue("term")
	scrapper.Scrape(term)
	return c.Attachment("jobs.csv", "job.csv")
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)

	e.Logger.Fatal(e.Start(":1323"))
	scrapper.Scrape("python")
}
