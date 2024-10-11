package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	contextRoot       = "localhost:8080"
	timeFormat        = time.TimeOnly
	contentType       = "text/plain"
	errorInvalidCT    = "Invalid Content-Type, expected text/plain"
	errorReadBody     = "Failed to read request body"
	errorTimeParse    = "Wrong time format"
	errorNotSupported = "Not supported method"
)

type TimeStampManager struct {
	updateCh chan time.Time
	readCh   chan chan time.Time
}

func NewTimeStampManager() *TimeStampManager {
	tsm := &TimeStampManager{
		updateCh: make(chan time.Time),
		readCh:   make(chan chan time.Time),
	}
	go tsm.run(time.Now())
	return tsm
}

func (tsm *TimeStampManager) run(initialTime time.Time) {
	currentTime := initialTime
	for {
		select {
		case newTime := <-tsm.updateCh:
			currentTime = newTime
		case replyCh := <-tsm.readCh:
			replyCh <- currentTime
		}
	}
}

func (tsm *TimeStampManager) UpdateTimeStamp(newTime time.Time) {
	tsm.updateCh <- newTime
}

func (tsm *TimeStampManager) GetTimeStamp() time.Time {
	replyCh := make(chan time.Time)
	tsm.readCh <- replyCh
	return <-replyCh
}

func errorResponse(writer http.ResponseWriter, message string, status int) {
	http.Error(writer, message, status)
}

func handleRootGET(writer http.ResponseWriter, request *http.Request, tsm *TimeStampManager) {
	if request.Header.Get("Content-Type") != contentType {
		errorResponse(writer, errorInvalidCT, http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", contentType)
	currentTime := tsm.GetTimeStamp()

	_, err := writer.Write([]byte(currentTime.Format(timeFormat)))
	if err != nil {
		log.Fatalf("Error writing response, %v", err)
	}
}

func handleRootPOST(writer http.ResponseWriter, request *http.Request, tsm *TimeStampManager) {
	if request.Header.Get("Content-Type") != contentType {
		errorResponse(writer, errorInvalidCT, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		errorResponse(writer, errorReadBody, http.StatusBadRequest)
		return
	}

	newTime, err := time.Parse(timeFormat, string(body))
	if err != nil {
		errorResponse(writer, errorTimeParse, http.StatusBadRequest)
		return
	}

	tsm.UpdateTimeStamp(newTime)
}

func runServer(tsm *TimeStampManager) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			handleRootGET(writer, request, tsm)
		} else if request.Method == http.MethodPost {
			handleRootPOST(writer, request, tsm)
		} else {
			errorResponse(writer, errorNotSupported, http.StatusBadRequest)
		}
	})

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
	req, err := http.NewRequest("GET", "http://"+contextRoot, nil)
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
	go runServer(NewTimeStampManager())

	postTimeStamp()

	getTimeStamp()

	select {}
}
