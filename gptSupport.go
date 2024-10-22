package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	//"github.com/docker/docker/api/types"
	con "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Replace this with your actual API key

// analyzeLogWithChatGpt sends logs to ChatGPT and gets an analysis
func analyzeLogWithChatGpt(logs string) (string, error) {
	messages := []Message{
		{Role: "user", Content: fmt.Sprintf("Analyze the following logs and provide potential fixes: %s", logs)},
	}
	requestBody := &Request{Model: "gpt-3.5-turbo", Messages: messages}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gptResponse Response
	err = json.Unmarshal(body, &gptResponse)
	if err != nil {
		return "", err
	}

	if len(gptResponse.Choices) > 0 {
		return gptResponse.Choices[0].Message.Content, nil
	}

	return "No analysis available", errors.New("Failed to get analysis")
}

// monitorContainer monitors Docker containers and analyzes logs when they exit or die
func monitorContainer() {
	log.Println("Starting monitoring")
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	for {
		log.Println("Getting container list")
		containers, err := cli.ContainerList(context.Background(), con.ListOptions{All: true})
		if err != nil {
			log.Fatalf("Failed to list Docker containers: %v", err)
		}

		for _, container := range containers {
			if container.State == "exited" || container.State == "dead" {
				containerID := container.ID[:10]

				// Get container logs
				log.Println("Getting logs for container %v", container)
				logs, err := cli.ContainerLogs(context.Background(), container.ID, con.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					log.Printf("Failed to get logs for container %s: %v", containerID, err)
					continue
				}
				logsData, err := ioutil.ReadAll(logs)
				if err != nil {
					log.Printf("Failed to read logs for container %s: %v", containerID, err)
					continue
				}

				// Send logs to ChatGPT for analysis
				analysis, err := analyzeLogWithChatGpt(string(logsData))
				if err != nil {
					log.Printf("Error analyzing logs with ChatGPT: %v", err)
					continue
				}
				log.Printf("Got CHAT GPT response %v", analysis)
				// Store the container status, logs, and ChatGPT response
				containerData[containerID] = ContainerStatus{
					ID:              containerID,
					Name:            container.Names[0],
					Status:          container.State,
					Logs:            string(logsData),
					ChatGPTResponse: analysis,
				}
			} else {
				containerID := container.ID[:10]
				containerData[containerID] = ContainerStatus{
					ID:              containerID,
					Name:            container.Names[0],
					Status:          container.State,
					Logs:            "",
					ChatGPTResponse: "",
				}
			}
		}

		// Polling interval
		time.Sleep(pollInterval)
	}
}

// getContainerStatuses returns container statuses as a JSON response
func getContainerStatuses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containerData)
}

// Request structure for ChatGPT API
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message structure for ChatGPT API
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response structure for ChatGPT API
type Response struct {
	Choices []Choice `json:"choices"`
}

// Choice structure for ChatGPT API
type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}
