package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/?name=Vlad", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	var handler = &Handler{}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `Parsed query-param with key "name": Vlad`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestListHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/upload", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))

	defer ts.Close()

	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  ts.URL,
	}
	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `name: test.txt; size: 5; ext: .txt
`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestListExtHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/upload?ext=png", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))

	defer ts.Close()

	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  ts.URL,
	}
	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := ``

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestUploadHandler(t *testing.T) {
	file, _ := os.Open("test.txt")
	defer file.Close()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file) //nolint:errcheck
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload", body)

	req.Header.Add("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))

	defer ts.Close()
	uploadHandler := &UploadHandler{ //nolint:wsl
		UploadDir: "upload",
		HostAddr:  ts.URL,
	}
	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := ts.URL

	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
