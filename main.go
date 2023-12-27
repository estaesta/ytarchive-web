package main

import (
	"net/http"

	"github.com/estaesta/ytarchive-web/handler"
	"github.com/estaesta/ytarchive-web/utils"
	"github.com/estaesta/ytarchive-web/view"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Static("/static", "assets")

	component := view.Index()

	e.GET("/", func(c echo.Context) error {
		return utils.Render(c, http.StatusOK, component)
	})

	e.POST("/archive", handler.PostArchive)

	e.Logger.Fatal(e.Start(":1323"))
}
