package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var originalNames = map[string]string{}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}

	return string(b)
}

type UploadHandler struct {
	HostAddr  string
	UploadDir string
}

const timeOut = 30
const randomness = 12

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		reqExt := "." + r.FormValue("ext")
		err := filepath.Walk(h.UploadDir, func(path string, info os.FileInfo, _ error) error {
			if _, err := os.Stat(h.UploadDir); os.IsNotExist(err) {
				return err
			}

			if info.IsDir() {
				return nil
			}

			ext := filepath.Ext(path)

			if reqExt != "." && reqExt != ext {
				return nil
			}

			name, ok := originalNames[info.Name()]
			if !ok {
				name = info.Name()
			}

			i := fmt.Sprintf("name: %s; size: %d; ext: %s", name, info.Size(), ext)

			fmt.Fprintln(w, i)

			return nil
		})

		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
		}
	case http.MethodPost:
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusBadRequest)

			return
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)

		if err != nil {
			http.Error(w, "Unable to read file", http.StatusBadRequest)

			return
		}

		ext := filepath.Ext(header.Filename)

		newFileName := randSeq(randomness) + ext

		filePath := h.UploadDir + "/" + newFileName

		originalNames[newFileName] = header.Filename

		err = ioutil.WriteFile(filePath, data, 0777) //nolint:gomnd,gosec

		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)

			return
		}

		link := h.HostAddr + "/" + newFileName

		req, err := http.NewRequest(http.MethodHead, link, nil)

		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to check file", http.StatusInternalServerError)

			return
		}

		cli := &http.Client{} //nolint:exhaustivestruct

		resp, err := cli.Do(req)

		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to check file", http.StatusInternalServerError)

			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Println(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)

			return
		}

		fmt.Fprintln(w, link)
	}
}

type Employee struct {
	Name   string  `json:"name" xml:"name"`
	Age    int     `json:"age" xml:"age"`
	Salary float32 `json:"salary" xml:"salary"`
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		name := r.FormValue("name")
		fmt.Fprintf(w, "Parsed query-param with key \"name\": %s", name)
	case http.MethodPost:
		var emp Employee

		ct := r.Header.Get("Content-Type")

		switch ct {
		case "application/json":
			defer r.Body.Close()

			if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
				http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)

				return
			}
		default:
			http.Error(w, "Unknown content type", http.StatusBadRequest)

			return
		}

		fmt.Fprintf(w, "Got a new employee!\nName: %s\nAge: %dy.o.\nSalary %0.2f\n", emp.Name, emp.Age, emp.Salary)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  "http://localhost:8080",
	}

	handler := &Handler{}

	http.Handle("/", handler)
	http.Handle("/upload", uploadHandler)

	srv := http.Server{ //nolint:exhaustivestruct
		Addr:              ":8000",
		ReadTimeout:       timeOut * time.Second,
		WriteTimeout:      timeOut * time.Second,
		ReadHeaderTimeout: timeOut * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)

			return
		}
	}()

	log.Println("main server started at localhost:8000")

	dirToServe := http.Dir(uploadHandler.UploadDir)

	fs := http.Server{ //nolint:exhaustivestruct
		Addr:              ":8080",
		Handler:           http.FileServer(dirToServe),
		ReadTimeout:       timeOut * time.Second,
		WriteTimeout:      timeOut * time.Second,
		ReadHeaderTimeout: timeOut * time.Second,
	}

	log.Println("file server started at localhost:8080")

	if err := fs.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
