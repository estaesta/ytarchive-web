package main

import (
	"net/http"

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
	e.Logger.Fatal(e.Start(":1323"))

	// http.Handle("/", templ.Handler(component))
	//
	// http.ListenAndServe(":1323", nil)
}
