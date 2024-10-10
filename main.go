package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var timeStamp time.Time

const (
	contextRoot    = "localhost:8080"
	timeFormat     = time.RFC3339
	contentType    = "text/plain"
	errorInvalidCT = "Invalid Content-Type, expected text/plain"
	errorReadBody  = "Failed to read request body"
	errorTimeParse = "Wrong time format"
)

func errorResponse(writer http.ResponseWriter, message string, status int) {
	http.Error(writer, message, status)
}

func handleRootGET(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != contentType {
		errorResponse(writer, errorInvalidCT, http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", contentType)
	_, _ = writer.Write([]byte(timeStamp.Format(timeFormat)))
}

func handleRootPOST(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != contentType {
		errorResponse(writer, errorInvalidCT, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		errorResponse(writer, errorReadBody, http.StatusBadRequest)
		return
	}

	timeStamp, err = time.Parse(timeFormat, string(body))
	if err != nil {
		errorResponse(writer, errorTimeParse, http.StatusBadRequest)
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

	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending POST request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("POST request returned an error. Status: %d\n", resp.StatusCode)
	}
}

func getTimeStamp() {
	req, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		log.Fatalf("Error creating GET request: %v", err)
		return
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending GET request: %v", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("GET request returned an error. Status: %d\n", resp.StatusCode)
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
