package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "myprivatenetwork/internal/xray"
)

var service *xray.Service

func main() {
    // Определяем путь к конфигурационному файлу
    configPath := getConfigPath()

    // Инициализируем сервис
    var err error
    service, err = xray.NewService(configPath)
    if err != nil {
        log.Fatalf("Failed to initialize service: %v", err)
    }

    // Настраиваем маршруты
    http.HandleFunc("/api/clients", handleClients)
    http.HandleFunc("/api/clients/create", handleCreateClient)

    // Запускаем сервер
    log.Printf("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

func getConfigPath() string {
    // Сначала проверяем переменную окружения
    if path := os.Getenv("XRAY_CONFIG_PATH"); path != "" {
        return path
    }
    // Иначе используем путь по умолчанию
    return "config.json"
}

func handleClients(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        clients := service.GetClients()
        json.NewEncoder(w).Encode(clients)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleCreateClient(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        Email string `json:"email"`
        Level int    `json:"level"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    config := xray.ClientConfig{
        Email: request.Email,
        Level: request.Level,
    }

    client, link, err := service.CreateClient(config)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := struct {
        Client *xray.Client `json:"client"`
        Link   string       `json:"link"`
    }{
        Client: client,
        Link:   link,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
} 