package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type gofileResponse struct {
	Status string `json:"status"`
	Data   struct {
		DownloadPage string `json:"downloadPage"`
		Code         string `json:"code"`
		ParentFolder string `json:"parentFolder"`
		FileID       string `json:"fileId"`
		FileName     string `json:"fileName"`
		MD5          string `json:"md5"`
	} `json:"data"`
}

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

func curlGofile(path string, server string) (gofileResponse, error) {
	var response gofileResponse
	// curl -F "file=@someFile" https://store1.gofile.io/upload
	// copied from curlconverter.com
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	fw, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}
	fd, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}

	writer.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://%s.gofile.io/uploadFile", server), form)
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}
	fmt.Printf("%s\n", bodyText)
	// unmarshal the response
	err = json.Unmarshal(bodyText, &response)
	if err != nil {
		log.Fatal(err)
		return gofileResponse{}, err
	}

	return response, nil
}

// Upload the downloaded directory to Gofile
func UploadToGofile(dirPath string) (string, error) {
	// remove the directory after uploading
	defer func() {
		err := os.RemoveAll(dirPath)
		if err != nil {
			fmt.Println("failed to remove directory")
		}
	}()

	// upload the file in the directory to Gofile
	// the directory only contains one file
	// the name of the file is unknown
	entry, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("failed to read directory")
		return "", err
	}

	// get the path of the file
	filePath := filepath.Join(dirPath, entry[0].Name())

	// upload the file to Gofile
	fmt.Println("uploading to Gofile")
	response, err := curlGofile(filePath, "store11")
	if err != nil {
		fmt.Println("failed to upload to Gofile")
		return "", err
	}
	// check if response is ok
	if response.Status != "ok" {
		fmt.Println("response is not ok")
		return "", fmt.Errorf("response is not ok")
	}

	downloadPage := response.Data.DownloadPage

	return downloadPage, nil
}

// Run yt-dlp to download the video concurrently
// save the stdout to a buffer channel
// ytarchive -v -o "archive/%(id)s/[[%(upload_date)s]_%(title)s(%(id)s)" --add-metadata -merge -w https://www.youtube.com/watch\?v\=videoId best
func DownloadVideo(url string, directory string, outchan chan string) {
	cmd := exec.Command("yt-dlp", "-o", directory+"/%(id)s/%(title)s.%(ext)s", url)
	// cmd := exec.Command("./counter")
	// cmd := exec.Command(
	// 	"ytarchive", "-v", "-o", directory+"/%(title)s.%(ext)s",
	// 	"--add-metadata", "-merge", "-w", url, "best")

	// defer close(outchan)
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
