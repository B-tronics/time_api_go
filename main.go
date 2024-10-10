package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var contextRoot = "localhost:8080"
var timeStamp time.Time
var timeFormat = time.RFC3339

func handleRootGET(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != "text/plain" {
		http.Error(writer, "Invalid Content-Type, expected text/plain", http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", "text/plain")
	_, _ = writer.Write([]byte(timeStamp.Format(timeFormat)))
}

func handleRootPOST(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != "text/plain" {
		http.Error(writer, "Invalid Content-Type, expected text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Failed to read request body", http.StatusBadRequest)
		return
	}

	timeStamp, err = time.Parse(timeFormat, string(body))
	if err != nil {
		http.Error(writer, "Wrong time format", http.StatusBadRequest)
		return
	}
}

func runServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRootGET)
	mux.HandleFunc("POST /", handleRootPOST)

	if err := http.ListenAndServe(contextRoot, mux); err != nil {
		log.Fatal(err.Error())
	}
}

func postTimeStamp() {
	body := []byte(time.Now().Format(timeFormat))
	req, err := http.NewRequest("POST", "http://"+contextRoot, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error creating POST request: %v", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending POST request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to send POST request. Status: %d\n", resp.StatusCode)
	}
}

func getTimeStamp() {
	req, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(body))
}

func main() {
	go runServer()

	postTimeStamp()

	getTimeStamp()

	select {}
}
