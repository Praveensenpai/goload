package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schollz/progressbar/v3"
)

type Download struct {
	ID          string  `json:"id"`
	URL         string  `json:"url"`
	Filename    string  `json:"filename"`
	Filepath    string  `json:"filepath"`
	Status      string  `json:"status"`
	SizeCurrent int64   `json:"size_current"`
	SizeTotal   int64   `json:"size_total"`
	Progress    float64 `json:"progress"`
	Speed       int64   `json:"speed"`
}

var (
	downloads   = make(map[string]*Download)
	downloadsMu sync.Mutex
	downloadDir = filepath.Join(os.Getenv("HOME"), "Downloads", "GoLoad")
)

func init() {
	os.MkdirAll(downloadDir, 0755)
	err := os.RemoveAll(downloadDir)
	if err != nil {
		log.Println("Error clearing GoLoad directory:", err)
	}
	os.MkdirAll(filepath.Join(downloadDir, "temp"), 0755)
}

func cleanInvalidDownloads() {
	downloadsMu.Lock()
	defer downloadsMu.Unlock()
	for id, dl := range downloads {
		if dl.Filepath != "" && dl.Status == "completed" {
			if _, err := os.Stat(dl.Filepath); os.IsNotExist(err) {
				delete(downloads, id)
			}
		}
	}
}

func addDownload(c *gin.Context) {
	cleanInvalidDownloads()
	var req struct {
		URL string `json:"url"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	downloadsMu.Lock()
	for _, dl := range downloads {
		if dl.URL == req.URL {
			c.JSON(http.StatusOK, gin.H{"message": "Download already exists", "id": dl.ID})
			downloadsMu.Unlock()
			return
		}
	}
	downloadsMu.Unlock()

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	downloadsMu.Lock()
	downloads[id] = &Download{ID: id, URL: req.URL, Status: "in_progress"}
	downloadsMu.Unlock()

	go startDownload(id, req.URL)

	c.JSON(http.StatusOK, gin.H{"message": "Download started", "id": id})
}

func startDownload(id, fileURL string) {
	resp, err := http.Get(fileURL)
	if err != nil {
		updateStatus(id, "failed")
		return
	}
	defer resp.Body.Close()

	filename := getFilename(resp, fileURL)
	categoryDir := getCategoryDir(filename, resp.Header.Get("Content-Type"))
	tempPath := filepath.Join(downloadDir, "temp", filename+".goloadtemp")
	finalPath := filepath.Join(categoryDir, filename)

	outFile, err := os.Create(tempPath)
	if err != nil {
		updateStatus(id, "failed")
		return
	}
	defer outFile.Close()

	sizeTotal := resp.ContentLength
	bar := progressbar.DefaultBytes(sizeTotal, "Downloading")
	downloadsMu.Lock()
	downloads[id].Filename = filename
	downloads[id].Filepath = finalPath
	downloads[id].SizeTotal = sizeTotal
	downloadsMu.Unlock()

	var sizeCurrent int64
	writer := io.MultiWriter(outFile, bar)
	buf := make([]byte, 32*1024)
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := writer.Write(buf[:n])
			if writeErr != nil {
				updateStatus(id, "failed")
				return
			}
			sizeCurrent += int64(n)
			duration := time.Since(startTime).Seconds()
			if duration > 0 {
				speed := int64(float64(sizeCurrent) / duration)
				downloadsMu.Lock()
				downloads[id].Speed = speed
				downloadsMu.Unlock()
			}
			progress := (float64(sizeCurrent) / float64(sizeTotal)) * 100

			downloadsMu.Lock()
			downloads[id].SizeCurrent = sizeCurrent
			downloads[id].Progress = progress
			downloadsMu.Unlock()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			updateStatus(id, "failed")
			return
		}
	}

	os.Rename(tempPath, finalPath)
	updateStatus(id, "completed")
}

func getFilename(resp *http.Response, fileURL string) string {
	parsedURL, _ := url.Parse(fileURL)
	filename := filepath.Base(parsedURL.Path)
	decodedFilename, err := url.QueryUnescape(filename)
	if err != nil {
		return filename
	}
	return decodedFilename
}

func getCategoryDir(filename, mimeType string) string {
	ext := filepath.Ext(filename)
	if mimeType == "" {
		mimeType = mime.TypeByExtension(ext)
	}
	category := "unknown"
	switch {
	case mimeType == "video/mp4" || mimeType == "video/x-matroska":
		category = "videos"
	case mimeType == "audio/mpeg" || mimeType == "audio/wav" || mimeType == "audio/flac":
		category = "audio"
	case mimeType == "application/zip" || mimeType == "application/x-rar-compressed":
		category = "compressed"
	}
	dir := filepath.Join(downloadDir, category)
	os.MkdirAll(dir, 0755)
	return dir
}

func updateStatus(id, status string) {
	downloadsMu.Lock()
	defer downloadsMu.Unlock()
	if dl, exists := downloads[id]; exists {
		dl.Status = status
	}
}

func getDownloads(c *gin.Context) {
	cleanInvalidDownloads()
	downloadsMu.Lock()
	defer downloadsMu.Unlock()

	var list []Download
	for _, dl := range downloads {
		list = append(list, *dl)
	}
	c.JSON(http.StatusOK, list)
}

func clearFailed(c *gin.Context) {
	cleanInvalidDownloads()
	downloadsMu.Lock()
	defer downloadsMu.Unlock()

	for id, dl := range downloads {
		if dl.Status == "failed" {
			delete(downloads, id)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Failed downloads cleared"})
}

func main() {
	const PORT uint16 = 6060
	r := gin.Default()
	r.POST("/add", addDownload)
	r.GET("/downloads", getDownloads)
	r.DELETE("/clear_failed", clearFailed)

	log.Printf("GoLoad server running on http://localhost:%d 🚀 \n", PORT)
	r.Run(fmt.Sprintf(":%d", PORT))
}
