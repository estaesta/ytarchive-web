package main

import (
	"fmt"
	"net/http"

	"github.com/estaesta/ytarchive-web/handler"
	"github.com/estaesta/ytarchive-web/utils"
	"github.com/estaesta/ytarchive-web/view"
	"github.com/labstack/echo/v4"

	// "github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)

	defer func() {
		err := nc.Drain()
		if err != nil {
			fmt.Println("failed to drain the connection")
		}
	}()

	e := echo.New()

	// e.Use(middleware.Logger())
	e.Static("/static", "assets")

	component := view.Index()

	e.GET("/", func(c echo.Context) error {
		return utils.Render(c, http.StatusOK, component)
	})

	postArchive := func(c echo.Context) error {
		return handler.PostArchive(c, nc)
	}
	e.POST("/archive", postArchive)

	getArchive := func(c echo.Context) error {
		return handler.GetArchive(c, nc)
	}
	e.GET("/archive/:videoId", getArchive)

	e.Logger.Fatal(e.Start(":1323"))
}
