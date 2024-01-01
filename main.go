package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/estaesta/ytarchive-web/handler"
	"github.com/estaesta/ytarchive-web/utils"
	"github.com/estaesta/ytarchive-web/view"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	e := echo.New()
	domain := os.Getenv("DOMAIN")
	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(domain)
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
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
	e.Use(middleware.HTTPSWWWRedirect())
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

	// e.Logger.Fatal(e.Start(":1323"))
	e.Logger.Fatal(e.StartAutoTLS(":443"))
}
