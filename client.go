package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"time"
)

type RequestConfig struct {
	Method  string
	URL     string
	Body    io.Reader
	Headers map[string]string
}

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
		Timeout: 10 * time.Second,
	}

	// Define URLs for different requests
	urls := []RequestConfig{
		{
			Method: "GET",
			URL:    "http://localhost:8443/",
		},
		{
			Method: "POST",
			URL:    "http://localhost:8443/postjson",
			Body: func() io.Reader {
				data := map[string]string{"message": "Hello, server!"}
				jsonData, _ := json.Marshal(data)
				return bytes.NewBuffer(jsonData)
			}(),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			Method: "POST",
			URL:    "http://localhost:8443/upload",
			Body: func() io.Reader {
				file, _ := os.Open("testfile.txt")
				defer file.Close()

				var b bytes.Buffer
				w := multipart.NewWriter(&b)
				fw, _ := w.CreateFormFile("uploadfile", "testfile.txt")
				_, _ = io.Copy(fw, file)
				w.Close()
				return &b
			}(),
			Headers: map[string]string{
				"Content-Type": "multipart/form-data",
			},
		},
	}

	// Make requests
	for _, reqConfig := range urls {
		response, err := makeRequest(client, reqConfig)
		if err != nil {
			fmt.Printf("Error making %s request: %v\n", reqConfig.Method, err)
			continue
		}
		fmt.Printf("%s response: %s\n", reqConfig.Method, response)
	}
}

func makeRequest(client *http.Client, reqConfig RequestConfig) (string, error) {
	req, err := http.NewRequest(reqConfig.Method, reqConfig.URL, reqConfig.Body)
	if err != nil {
		return "", err
	}

	for k, v := range reqConfig.Headers {
		req.Header.Set(k, v)
	}

	// Implementing context for timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
