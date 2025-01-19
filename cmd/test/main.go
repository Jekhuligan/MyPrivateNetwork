package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
)

const (
    API_URL  = "http://116.203.117.243:8080"
    USERNAME = "tay"
    PASSWORD = "123"
)

type CreateClientRequest struct {
    Email string `json:"email"`
    Level int    `json:"level"`
}

type Client struct {
    ID      string `json:"id"`
    Email   string `json:"email"`
    Level   int    `json:"level"`
    AlterId int    `json:"alterId"`
}

type CreateClientResponse struct {
    Client *Client `json:"client"`
    Link   string  `json:"link"`
}

func main() {
    // Создаем запрос на создание клиента
    reqBody := CreateClientRequest{
        Email: "User@net.com",
        Level: 1,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        log.Fatalf("Failed to marshal request: %v", err)
    }

    // Создаем HTTP запрос
    req, err := http.NewRequest("POST", API_URL+"/api/clients/create", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatalf("Failed to create request: %v", err)
    }

    // Добавляем заголовки
    req.Header.Set("Content-Type", "application/json")
    req.SetBasicAuth(USERNAME, PASSWORD)

    // Выполняем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Failed to send request: %v", err)
    }
    defer resp.Body.Close()

    // Читаем ответ
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("Failed to read response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        log.Fatalf("API request failed with status %d: %s", resp.StatusCode, string(body))
    }

    // Парсим ответ
    var response CreateClientResponse
    if err := json.Unmarshal(body, &response); err != nil {
        log.Fatalf("Failed to parse response: %v", err)
    }

    // Выводим результат
    fmt.Println("Client created successfully!")
    fmt.Printf("Client ID: %s\n", response.Client.ID)
    fmt.Printf("Client Email: %s\n", response.Client.Email)
    fmt.Printf("Connection Link: %s\n", response.Link)
} 