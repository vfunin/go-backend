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
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `Parsed query-param with key "name": Vlad`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestUploadHandler(t *testing.T) {
	// открываем файл, который хотим отправить
	file, _ := os.Open("test.txt")
	defer file.Close()
	// действия, необходимые для того, чтобы засунуть файл в запрос
	// в качестве мультипарт-формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file) //nolint:errcheck
	writer.Close()
	// опять создаем запрос, теперь уже на /upload эндпоинт
	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	// создаем ResponseRecorder
	rr := httptest.NewRecorder()
	// создаем заглушку файлового сервера. Для прохождения тестов
	// нам достаточно чтобы он возвращал 200 статус
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()
	uploadHandler := &UploadHandler{ //nolint:wsl
		UploadDir: "upload",
		// таким образом мы подменим адрес файлового сервера
		// и вместо реального, хэндлер будет стучаться на заглушку
		// которая всегда будет возвращать 200 статус, что нам и нужна
		HostAddr: ts.URL,
	}
	// опять же, вызываем ServeHTTP у тестируемого обработчика
	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := ts.URL + "/test.txt"

	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
