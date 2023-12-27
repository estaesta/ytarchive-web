package utils

import (
	"context"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Render(ctx echo.Context, status int, t templ.Component) error {
    ctx.Response().Writer.WriteHeader(status)

    err := t.Render(context.Background(), ctx.Response().Writer)
    if err != nil {
	return ctx.String(http.StatusInternalServerError, "failed to render response template")
    }

    return nil
}

// Upload the downloaded directory to Gofile
func UploadToGofile(directory string) error {
    // remove the directory after uploading
    defer os.RemoveAll(directory)

    //TODO: upload the directory to Gofile using the API

    return nil
}
