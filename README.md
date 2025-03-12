# GoLoad

[![Go Version](https://img.shields.io/badge/Go-1.16+-blue.svg)](https://golang.org) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

GoLoad is a lightweight, concurrent download manager written in Go. It provides a simple RESTful API for downloading files from a given URL while automatically tracking progress and categorizing files based on their MIME type.

## Features

- **Concurrent Downloads:** Handle multiple downloads simultaneously with thread-safe operations.
- **Real-Time Progress:** Visual progress bar integration using [schollz/progressbar](https://github.com/schollz/progressbar).
- **File Categorization:** Automatically organizes downloads into directories (e.g., videos, audio, compressed, unknown) based on file type.
- **RESTful API:** Easily interact with the application using a set of intuitive endpoints powered by [Gin](https://github.com/gin-gonic/gin).
- **Automatic Cleanup:** Invalid or failed downloads are cleaned up to keep your download list current.

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.16 or later

### Steps

1. **Clone the Repository:**

   ```bash
   git clone https://github.com/Praveensenpai/goload.git
   cd goload
   ```

2. **Download Dependencies:**

   ```bash
   go mod download
   ```

3. **Build the Application:**

   ```bash
   go build -o goload
   ```

4. **Run the Server:**

   ```bash
   ./goload
   ```

   The server will start on [http://localhost:6060](http://localhost:6060).

## API Endpoints

### Start a Download

- **Endpoint:** `POST /add`
- **Description:** Initiates a new download.
- **Request Body:**

  ```json
  {
    "url": "https://example.com/path/to/file.mp4"
  }
  ```

- **Response:**

  ```json
  {
    "message": "Download started",
    "id": "unique_download_id"
  }
  ```

*Note: If the download already exists, you'll receive a message indicating so along with the existing download ID.*

### List Downloads

- **Endpoint:** `GET /downloads`
- **Description:** Retrieves a list of all current downloads with details including status, progress, and speed.
- **Response Example:**

  ```json
  [
    {
      "id": "unique_download_id",
      "url": "https://example.com/path/to/file.mp4",
      "filename": "file.mp4",
      "filepath": "/home/user/Downloads/GoLoad/videos/file.mp4",
      "status": "in_progress",
      "size_current": 1048576,
      "size_total": 2097152,
      "progress": 50.0,
      "speed": 51200
    }
  ]
  ```

### Clear Failed Downloads

- **Endpoint:** `DELETE /clear_failed`
- **Description:** Clears downloads that have failed.
- **Response:**

  ```json
  {
    "message": "Failed downloads cleared"
  }
  ```

## How It Works

- **Download Flow:** When a new download is added, GoLoad creates a temporary file in `$HOME/Downloads/GoLoad/temp`. Once the download completes, the file is moved to a categorized directory (like videos or audio) based on its MIME type.
- **Progress Tracking:** A progress bar shows the real-time download progress, while download details (like current size, total size, and speed) are updated continuously.
- **Concurrency:** A mutex ensures thread-safe access to shared download data across concurrent downloads.

## Configuration

- **Download Directory:** Files are stored in the `GoLoad` directory inside your Downloads folder. This can be modified in the source code.
- **Server Port:** The server listens on port `6060` by default. Change the port number in `main.go` if needed.

## Contributing

Contributions are welcome! If you have ideas for improvements or encounter issues, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Enjoy seamless downloading with GoLoad! If you have any questions or suggestions, feel free to reach out.
