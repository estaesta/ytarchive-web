package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/estaesta/ytarchive-web/handler"
	"github.com/estaesta/ytarchive-web/utils"
	"github.com/estaesta/ytarchive-web/view"
	"github.com/labstack/echo/v4"

	// "github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	e := echo.New()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Println("failed to connect to nats server")
	}
	defer func() {
		err := nc.Drain()
		if err != nil {
			fmt.Println("failed to drain the connection")
		}
	}()

	js, _ := jetstream.New(nc)
	ctx := context.Background()

	// create kv store
	kv, _ := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: "videoStatus",
	})

	// e.Use(middleware.Logger())
	e.Static("/static", "assets")

	component := view.Index()

	e.GET("/", func(c echo.Context) error {
		return utils.Render(c, http.StatusOK, component)
	})

	postArchive := func(c echo.Context) error {
		return handler.PostArchive(c, nc, kv, ctx)
	}
	e.POST("/archive", postArchive)

	getArchive := func(c echo.Context) error {
		return handler.GetArchive(c, nc, kv, ctx)
	}
	e.GET("/archive/:videoId", getArchive)

	e.Logger.Fatal(e.Start(":1323"))
}
