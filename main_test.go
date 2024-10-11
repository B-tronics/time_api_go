package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewTimeStampManager(t *testing.T) {
	tsm := NewTimeStampManager()
	if tsm == nil {
		t.Error("Expected a non-nil TimeStampManager")
	}
}

func TestUpdateTimeStamp(t *testing.T) {
	tsm := NewTimeStampManager()
	newTime := time.Now()
	tsm.UpdateTimeStamp(newTime)
	if got := tsm.GetTimeStamp(); !got.Equal(newTime) {
		t.Errorf("Expected %v, got %v", newTime, got)
	}
}

func TestHandleRootGET(t *testing.T) {
	tsm := NewTimeStampManager()
	tsm.UpdateTimeStamp(time.Now())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	handleRootGET(rr, req, tsm)

	resp := rr.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	expected := tsm.GetTimeStamp().Format(timeFormat)
	if string(body) != expected {
		t.Errorf("Expected body %q, got %q", expected, string(body))
	}
}

func TestHandleRootGETInvalidCT(t *testing.T) {
	tsm := NewTimeStampManager()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handleRootGET(rr, req, tsm)

	resp := rr.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandleRootPOST(t *testing.T) {
	tsm := NewTimeStampManager()
	newTime := time.Now().Format(timeFormat)

	body := []byte(newTime)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	handleRootPOST(rr, req, tsm)

	if rr.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Result().StatusCode)
	}

	actual := tsm.GetTimeStamp().Format(timeFormat)

	if actual != newTime {
		t.Errorf("Expected timestamp %v, got %v", newTime, actual)
	}
}

func TestHandleRootPOSTInvalidCT(t *testing.T) {
	tsm := NewTimeStampManager()

	body := []byte(time.Now().Format(timeFormat))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handleRootPOST(rr, req, tsm)

	if rr.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Result().StatusCode)
	}
}

func TestHandleRootPOSTInvalidTimeFormat(t *testing.T) {
	tsm := NewTimeStampManager()

	body := []byte("invalid time format")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	handleRootPOST(rr, req, tsm)

	if rr.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Result().StatusCode)
	}
}

func TestServerIntegration(t *testing.T) {
	tsm := NewTimeStampManager()
	go runServer(tsm)

	time.Sleep(100 * time.Millisecond) // Allow server to start

	// Test POST
	newTime := time.Now().Format(timeFormat)
	body := []byte(newTime)
	req, err := http.NewRequest("POST", "http://"+contextRoot, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error creating POST request: %v", err)
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending POST request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test GET
	req, err = http.NewRequest("GET", "http://"+contextRoot, nil)
	if err != nil {
		log.Fatalf("Error creating GET request: %v", err)
		return
	}
	req.Header.Set("Content-Type", contentType)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending GET request: %v", err)
		return
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	expected := tsm.GetTimeStamp().Format(timeFormat)
	if string(bodyBytes) != expected {
		t.Errorf("Expected body %q, got %q", expected, string(bodyBytes))
	}
}

func TestConcurrentAccess(t *testing.T) {
	tsm := NewTimeStampManager()

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent writes
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			tsm.UpdateTimeStamp(time.Now().Add(time.Duration(i) * time.Second))
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Concurrent reads
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			got := tsm.GetTimeStamp()
			if got.IsZero() {
				t.Error("Received zero time")
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()
}
