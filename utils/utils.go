package utils

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Render(ctx echo.Context, status int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(status)

	err := t.Render(context.Background(), ctx.Response().Writer)
	if err != nil {
		return ctx.String(
			http.StatusInternalServerError,
			"failed to render response template",
		)
	}

	return nil
}

func RenderStream(ctx echo.Context, t templ.Component, event string) error {
	html, err := templ.ToGoHTML(context.Background(), t)
	if err != nil {
		return ctx.String(
			http.StatusInternalServerError,
			"failed to render response template",
		)
	}

	fmt.Fprintf(ctx.Response().Writer,
		"event: %s\ndata: %s\n\n", event, html)

	return nil
}

// Upload the downloaded directory to Gofile
func UploadToGofile(path string) error {
	// remove the directory after uploading
	// defer os.RemoveAll(path)

	// TODO: upload the directory to Gofile using the API

	return nil
}

// Run yt-dlp to download the video concurrently
// save the stdout to a buffer channel
func DownloadVideo(url string, directory string) chan string {
	outchan := make(chan string, 1)

	// execute yt-dlp using goroutine
	go func() {
		// cmd := exec.Command("yt-dlp", "-o", directory+"/%(title)s.%(ext)s", url)
		cmd := exec.Command("./counter")
		// cmd := exec.Command(
		// 	"ytarchive", "-v", "-o", directory+"/%(title)s.%(ext)s",
		// 	"--add-metadata", "-merge", "-w", url, "best")

		defer close(outchan)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println("failed to get stdout pipe")
		}

		err = cmd.Start()
		if err != nil {
			fmt.Println("failed to execute yt-dlp")
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Split(SplitFunc)
		for scanner.Scan() {
			outchan <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("error reading from scanner:", err)
		}

		fmt.Println("finished reading stdout")

		err = cmd.Wait()
		if err != nil {
			fmt.Println("failed to wait for yt-dlp")
		}

		err = stdout.Close()
		if err != nil {
			fmt.Println("failed to close stdout")
		}

		outchan <- "finished downloading"
	}()
	return outchan
}

// Parse the url to get the video id
// yt format is https://www.youtube.com/watch?v=videoID
// or https://youtu.be/videoID
func ParseYtURL(youtubeUrl string) (string, error) {
	u, err := url.Parse(youtubeUrl)
	if err != nil {
		return "", err
	}

	switch u.Host {
	case "youtu.be":
		return u.Path[1:], nil
	case "www.youtube.com":
		query := u.Query()
		videoID := query.Get("v")
		return videoID, nil
	default:
		return "", fmt.Errorf("invalid youtube url")
	}
}

// split fuction to split \r or \n
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 { // end of file
		return 0, nil, nil
	}

	// if we have a \r or \n, return the line
	if i := strings.Index(string(data), "\r"); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if i := strings.Index(string(data), "\n"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// if we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
