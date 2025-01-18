package xray

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
)

type Service struct {
    configPath string
    config     *XrayConfig
    serverAddr string
    serverPort int
}

func NewService(configPath string) (*Service, error) {
    service := &Service{
        configPath: configPath,
        config:     &XrayConfig{},
        serverAddr: "116.203.117.243", // Ваш адрес сервера
        serverPort: 30249,             // Ваш порт
    }
    
    // Загружаем существующую конфигурацию
    if err := service.loadConfig(); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    return service, nil
}

// loadConfig загружает конфигурацию из файла
func (s *Service) loadConfig() error {
    data, err := os.ReadFile(s.configPath)
    if err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    if err := json.Unmarshal(data, &s.config); err != nil {
        return fmt.Errorf("failed to parse config: %w", err)
    }

    return nil
}

// saveConfig сохраняет конфигурацию в файл
func (s *Service) saveConfig() error {
    data, err := json.MarshalIndent(s.config, "", "    ")
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    if err := os.WriteFile(s.configPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }

    return nil
}

// CreateClient создает нового клиента и добавляет его в конфигурацию
func (s *Service) CreateClient(config ClientConfig) (*Client, string, error) {
    // Создаем нового клиента
    client, err := NewClient(config)
    if err != nil {
        return nil, "", fmt.Errorf("failed to create client: %w", err)
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
        return nil, "", fmt.Errorf("failed to save config: %w", err)
    }

    // Перезапускаем Xray
    if err := s.restart(); err != nil {
        return nil, "", fmt.Errorf("failed to restart xray: %w", err)
    }

    // Генерируем ссылку для подключения
    link, err := GenerateClientLink(client, s.serverAddr, s.serverPort)
    if err != nil {
        return nil, "", fmt.Errorf("failed to generate client link: %w", err)
    }

    return client, link, nil
}

// GetConfig возвращает текущую конфигурацию
func (s *Service) GetConfig() *XrayConfig {
    return s.config
}

func (s *Service) restart() error {
    cmd := exec.Command("systemctl", "restart", "xray")
    return cmd.Run()
} 