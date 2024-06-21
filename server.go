package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type jsonData struct {
	msg string `json:"message"`
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Welcome to the Go HTTP Server!")
}

func handleJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method!= http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data jsonData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Received JSON: %s", data.msg)
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method!= http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("uploadfile")
	if err!= nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads/"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := os.Create(filepath.Join(uploadDir, header.Filename))
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s", header.Filename)
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/postjson", handleJSON)
	http.HandleFunc("/upload", handleFileUpload)

	fmt.Println("Starting server on :8443...")
	if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil); err!= nil {
		fmt.Println(err)
	}
}