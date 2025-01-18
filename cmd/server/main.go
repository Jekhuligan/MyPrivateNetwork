package main

import (
    "encoding/json"
    "log"
    "net/http"

    "myprivatenetwork/internal/xray"
)

var service *xray.Service

func main() {
    var err error
    service, err = xray.NewService("config.json")
    if err != nil {
        log.Fatalf("Failed to initialize service: %v", err)
    }

    http.HandleFunc("/api/clients", handleClients)
    http.HandleFunc("/api/clients/create", handleCreateClient)

    log.Printf("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

func handleClients(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        if len(service.GetConfig().Inbounds) > 0 {
            clients := service.GetConfig().Inbounds[0].Settings.Clients
            json.NewEncoder(w).Encode(clients)
        } else {
            json.NewEncoder(w).Encode([]xray.Client{})
        }
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
