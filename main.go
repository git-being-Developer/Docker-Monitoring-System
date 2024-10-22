package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	pollInterval = 10 * time.Second
)

var (
	apiKey    string
	orgId     string
	projectId string
)

var containerData = map[string]ContainerStatus{}

// struct to hold container data
type ContainerStatus struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	Logs            string `json:"logs"`
	ChatGPTResponse string `json:"chatgpt_response"`
}

func main() {

	intializeConifig()

	go monitorContainer()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	// Start HTTP server to serve container statuses
	http.HandleFunc("/containers", getContainerStatuses)

	fmt.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func intitializeConfig() {
	apiKey = os.Getenv("API_KEY")
	orgId = os.Getevne("ORGANISATION_ID")
	projectId = os.Getenv("PROJECT_ID")
	os.Setenv("DOCKER_API_VERSION", "1.45")
}
