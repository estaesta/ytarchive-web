package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/estaesta/ytarchive-web/utils"
	"github.com/estaesta/ytarchive-web/view"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func PostArchive(c echo.Context, nc *nats.Conn, kv jetstream.KeyValue, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// get the url from the form
	url := c.FormValue("yt-url")
	if url == "" {
		fmt.Println("url is empty")
		return c.String(http.StatusBadRequest, "url is empty")
	}
	fmt.Println(url)

	// parse the url to get the video id
	videoID, err := utils.ParseYtURL(url)
	if err != nil {
		fmt.Println("failed to parse url")
		return c.String(http.StatusBadRequest, "failed to parse url")
	}

	// if value is already in the kv store with value other than "downloading", return the url to the client
	value, err := kv.Get(ctx, "id."+videoID)
	if err != nil {
		fmt.Println(err)
	}
	if value != nil && string(value.Value()) != "downloading" {
		fmt.Println("video is already downloaded")
		return utils.Render(c, http.StatusOK, view.CloseSse("https://gofile.io/d/123456"))
	}

	// if the video id is already in the kv store, return the url to the client
	_, err = kv.Create(ctx, "id."+videoID, []byte("downloading"))
	if err != nil {
		fmt.Println(err)
		return utils.Render(c, http.StatusOK, view.CommandOutputHx(videoID))
	}

	fmt.Println("publishing to the topic:", videoID)

	outchan := make(chan string, 1)

	// execute yt-dlp using goroutine
	outchan = utils.DownloadVideo(url, "downloads")

	// TODO: upload the downloaded directory to Gofile using the API
	// use dummy api for now

	go func() {
		defer func() {
			// mu.Lock()
			// defer mu.Unlock()
			// delete(keyStatusMap, videoID)
		}()
		for msg := range outchan {
			err := nc.Publish(videoID, []byte(msg))
			if err != nil {
				fmt.Println("failed to publish to the topic" + videoID)
			}
		}
	}()

	// return utils.Render(c, http.StatusOK, view.Dummy(url))
	return utils.Render(c, http.StatusOK, view.CommandOutputHx(videoID))
}

func GetArchive(c echo.Context, nc *nats.Conn) error {
	videoID := c.Param("videoId")
	if videoID == "" {
		fmt.Println("video id is empty")
		return c.String(http.StatusBadRequest, "video id is empty")
	}

	// subscribe to the topic of the url and sent sse to the client
	msgChan := make(chan *nats.Msg)
	sub, err := nc.ChanSubscribe(videoID, msgChan)
	if err != nil {
		fmt.Println("failed to subscribe to the topic")
		return c.String(
			http.StatusInternalServerError,
			"failed to subscribe to the topic",
		)
	}
	fmt.Println("subscribed to the topic:", videoID)
	// defer sub.Unsubscribe()
	defer func() {
		err := sub.Unsubscribe()
		if err != nil {
			fmt.Println("failed to unsubscribe")
		}
	}()
	defer close(msgChan)

	// send sse to the client
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	fmt.Println("waiting for message")

breakLoop:
	for {
		select {
		case msg := <-msgChan:
			data := msg.Data
			if string(data) == "finished downloading" {
				break breakLoop
			}
			event := "archive-update"
			fmt.Fprintf(c.Response().Writer, "event: %s\ndata: %s\n\n", event, data)
			c.Response().Flush()
		case <-c.Request().Context().Done():
			fmt.Println("client disconnected")
			return nil
		}
	}

	fmt.Println("finished downloading")
	// path := fmt.Sprintf("downloads/%s", videoID)
	// go utils.UploadToGofile(path)

	// dummy upload by sending progress to the client
	for i := 0; i < 100; i++ {
		event := "archive-update"
		data := fmt.Sprintf("dummy uploading to Gofile: %d%%", i)
		fmt.Fprintf(c.Response().Writer, "event: %s\ndata: %s\n\n", event, data)
		c.Response().Flush()
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("finished uploading to Gofile")
	url := "https://gofile.io/d/123456"
	return utils.RenderStream(c, view.CloseSse(url), "archive-update")
}
