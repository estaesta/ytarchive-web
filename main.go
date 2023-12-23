package main

import (
	"net/http"

	"github.com/a-h/templ"
	// "github.com/labstack/echo/v4"
	// "github.com/labstack/echo/v4/middleware"
	"github.com/estaesta/ytarchive-web/view"
)

func main() {
	// e := echo.New()

	// e.Use(middleware.Logger())

	component := view.Index()

	// e.GET("/", func(c echo.Context) error {
	// 	return c.String(http.StatusOK, "Hello, World!")
	// })
	// e.Logger.Fatal(e.Start(":1323"))

	http.Handle("/", templ.Handler(component))

	http.ListenAndServe(":1323", nil)
}
