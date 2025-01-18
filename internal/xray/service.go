package xray

import (
    "fmt"
    "os/exec"
    "github.com/google/uuid"
)

type Service struct {
    configPath string
    config     *XrayConfig
}

func NewService(configPath string) *Service {
    return &Service{
        configPath: configPath,
        config:     &XrayConfig{},
    }
}

func (s *Service) generateUUID() string {
    return uuid.New().String()
}

func (s *Service) saveConfig() error {
    return s.config.SaveToFile(s.configPath)
}

func (s *Service) AddClient(email string) (*Client, error) {
    client := &Client{
        ID:      s.generateUUID(),
        Email:   email,
        Level:   1,
        AlterId: 0,
    }
    
    // Добавляем клиента в конфигурацию
    if len(s.config.Inbounds) > 0 {
        s.config.Inbounds[0].Settings.Clients = append(
            s.config.Inbounds[0].Settings.Clients,
            *client,
        )
    }
    
    // Сохраняем конфигурацию
    if err := s.saveConfig(); err != nil {
        return nil, err
    }
    
    // Перезапускаем Xray
    if err := s.restart(); err != nil {
        return nil, err
    }
    
    return client, nil
}

func (s *Service) RemoveClient(email string) error {
    if len(s.config.Inbounds) == 0 {
        return fmt.Errorf("no inbounds configured")
    }
    
    clients := s.config.Inbounds[0].Settings.Clients
    for i, client := range clients {
        if client.Email == email {
            // Удаляем клиента
            s.config.Inbounds[0].Settings.Clients = append(
                clients[:i],
                clients[i+1:]...,
            )
            
            // Сохраняем конфигурацию
            if err := s.saveConfig(); err != nil {
                return err
            }
            
            // Перезапускаем Xray
            return s.restart()
        }
    }
    
    return fmt.Errorf("client not found")
}

func (s *Service) restart() error {
    cmd := exec.Command("systemctl", "restart", "xray")
    return cmd.Run()
}

// CreateClient создает нового клиента и добавляет его в конфигурацию
func (s *Service) CreateClient(config ClientConfig) (*Client, string, error) {
    client := &Client{
        ID:    s.generateUUID(),
        Email: config.Email,
        // ... другие поля
    }
    
    // ... логика создания клиента
    
    if err := s.saveConfig(); err != nil {
        return nil, "", err
    }
    
    return client, s.generateLink(client), nil
}

func (s *Service) generateLink(client *Client) string {
    // Используйте правильное поле из конфига
    serverAddr := s.config.Address // или как у вас называется поле адреса сервера
    // ... генерация ссылки
    return ""
} 