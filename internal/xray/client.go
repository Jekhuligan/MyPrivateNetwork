package xray

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// ClientConfig содержит настройки для создания нового клиента
type ClientConfig struct {
    Email     string
    Level     int
    AlterId   int
    ExpiresAt *time.Time
}

// NewClient создает нового клиента с заданными параметрами
func NewClient(config ClientConfig) (*Client, error) {
    if config.Email == "" {
        return nil, fmt.Errorf("email is required")
    }

    // Генерируем UUID для клиента
    id, err := uuid.NewRandom()
    if err != nil {
        return nil, fmt.Errorf("failed to generate UUID: %w", err)
    }

    // Если уровень не указан, используем значение по умолчанию
    if config.Level == 0 {
        config.Level = 1
    }

    client := &Client{
        ID:      id.String(),
        Email:   config.Email,
        Level:   config.Level,
        AlterId: config.AlterId,
    }

    return client, nil
}

// GenerateClientLink генерирует ссылку для подключения клиента
func GenerateClientLink(client *Client, server string, port int) (string, error) {
    // Создаем базовую конфигурацию для ссылки
    config := map[string]interface{}{
        "v":    "2",
        "ps":   client.Email,
        "add":  server,
        "port": port,
        "id":   client.ID,
        "aid":  client.AlterId,
        "net":  "tcp",
        "type": "none",
        "sni":  "",
    }

    // Преобразуем конфигурацию в JSON
    configJSON, err := json.Marshal(config)
    if err != nil {
        return "", fmt.Errorf("failed to marshal config: %w", err)
    }

    // Кодируем конфигурацию в base64
    encoded := base64.StdEncoding.EncodeToString(configJSON)

    // Формируем итоговую ссылку
    link := fmt.Sprintf("vmess://%s", encoded)

    return link, nil
}

func (c *Client) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        ID    string `json:"id"`
        Email string `json:"email"`
    }{
        ID:    c.ID,
        Email: c.Email,
    })
} 