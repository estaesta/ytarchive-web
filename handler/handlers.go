package handler

import (
	"fmt"
	"net/http"

	"github.com/estaesta/ytarchive-web/utils"
	"github.com/estaesta/ytarchive-web/view"
	"github.com/labstack/echo/v4"
)

func PostArchive(c echo.Context) error {
	url := c.FormValue("yt-url")
	if url == "" {
		fmt.Println("url is empty")
		return c.String(http.StatusBadRequest, "url is empty")
	}
	fmt.Println(url)

	return utils.Render(c, http.StatusOK, view.Dummy(url))
}
