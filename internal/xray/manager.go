package xray

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type XrayManager struct {
	client    *http.Client
	serverURL string
	username  string
	password  string
}

// Добавляем структуру для информации о клиенте
type ClientInfo struct {
	ID         string
	Email      string
	ExpiryTime time.Time
	Enable     bool
	Port       int
	Up         int64
	Down       int64
	Total      int64
	LimitIP    int
	TotalGB    int64
	Flow       string
}

func NewXrayManager(serverURL, username, password string) (*XrayManager, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
		Jar:       jar,
	}

	manager := &XrayManager{
		client:    client,
		serverURL: serverURL,
		username:  username,
		password:  password,
	}

	if err := manager.login(); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return manager, nil
}

func (m *XrayManager) login() error {
	loginReq := LoginRequest{
		Username: m.username,
		Password: m.password,
	}

	loginData, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %v", err)
	}

	resp, err := m.client.Post(m.serverURL+"/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		return fmt.Errorf("failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	// Получаем страницу панели для установки куки
	_, err = m.client.Get(m.serverURL + "/panel")
	if err != nil {
		return fmt.Errorf("failed to get panel page: %v", err)
	}

	return nil
}

func (m *XrayManager) DeleteClientsByEmail(email string) error {
	// Получаем список inbounds
	listReq, err := http.NewRequest("GET", m.serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return fmt.Errorf("failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := m.client.Do(listReq)
	if err != nil {
		return fmt.Errorf("failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp InboundResponse
	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	for _, inbound := range inboundResp.Obj {
		var settings VlessSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		var clientsUpdated bool
		var newClients []VlessClient
		for _, c := range settings.Clients {
			if c.Email != email {
				newClients = append(newClients, c)
			} else {
				clientsUpdated = true
			}
		}

		if clientsUpdated {
			settings.Clients = newClients
			newSettings, err := json.Marshal(settings)
			if err != nil {
				return fmt.Errorf("failed to marshal new settings: %v", err)
			}
			inbound.Settings = string(newSettings)

			updateData, err := json.Marshal(inbound)
			if err != nil {
				return fmt.Errorf("failed to marshal update request: %v", err)
			}

			updateReq, err := http.NewRequest("POST",
				fmt.Sprintf("%s/panel/inbound/update/%d", m.serverURL, inbound.ID),
				bytes.NewBuffer(updateData))
			if err != nil {
				return fmt.Errorf("failed to create update request: %v", err)
			}

			updateReq.Header.Set("Content-Type", "application/json")
			updateReq.Header.Set("X-Requested-With", "XMLHttpRequest")

			updateResp, err := m.client.Do(updateReq)
			if err != nil {
				return fmt.Errorf("failed to send update request: %v", err)
			}
			defer updateResp.Body.Close()
		}
	}

	return nil
}

func (m *XrayManager) ClientExists(email string) (bool, error) {
	// Проверяем логин перед выполнением операции
	if err := m.checkAndRelogin(); err != nil {
		return false, fmt.Errorf("failed to check login status: %v", err)
	}

	listReq, err := http.NewRequest("GET", m.serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return false, fmt.Errorf("failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := m.client.Do(listReq)
	if err != nil {
		return false, fmt.Errorf("failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp InboundResponse
	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %v", err)
	}

	for _, inbound := range inboundResp.Obj {
		var settings VlessSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		for _, client := range settings.Clients {
			if client.Email == email {
				return true, nil
			}
		}
	}

	return false, nil
}

func (m *XrayManager) CreateClient(email string) (string, error) {
	// Проверяем логин перед выполнением операции
	if err := m.checkAndRelogin(); err != nil {
		return "", fmt.Errorf("failed to check login status: %v", err)
	}

	// Проверяем существование клиента
	exists, err := m.ClientExists(email)
	if err != nil {
		return "", fmt.Errorf("failed to check client existence: %v", err)
	}

	// Если клиент существует, удаляем его
	if exists {
		if err := m.DeleteClientsByEmail(email); err != nil {
			return "", fmt.Errorf("failed to delete existing client: %v", err)
		}
	}

	// Получаем список inbounds
	listReq, err := http.NewRequest("GET", m.serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := m.client.Do(listReq)
	if err != nil {
		return "", fmt.Errorf("failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp InboundResponse
	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		return "", fmt.Errorf("failed to decode inbounds response: %v", err)
	}

	// Ищем нужный inbound
	var targetInbound *Inbound
	for i := range inboundResp.Obj {
		if inboundResp.Obj[i].ID == 3 {
			targetInbound = &inboundResp.Obj[i]
			break
		}
	}

	if targetInbound == nil {
		return "", fmt.Errorf("inbound with ID=3 not found")
	}

	// Создаем нового клиента
	expiryTime := time.Now().AddDate(0, 1, 0).UnixMilli()
	newClient := VlessClient{
		ID:         generateUUID(),
		Flow:       "xtls-rprx-vision",
		Email:      email,
		LimitIP:    0,
		TotalGB:    0,
		ExpiryTime: expiryTime,
		Enable:     true,
	}

	// Обновляем настройки
	var settings VlessSettings
	if err := json.Unmarshal([]byte(targetInbound.Settings), &settings); err != nil {
		return "", fmt.Errorf("failed to parse settings: %v", err)
	}

	settings.Clients = append(settings.Clients, newClient)
	newSettings, err := json.Marshal(settings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal new settings: %v", err)
	}
	targetInbound.Settings = string(newSettings)

	// Отправляем запрос на обновление
	updateData, err := json.Marshal(targetInbound)
	if err != nil {
		return "", fmt.Errorf("failed to marshal update request: %v", err)
	}

	updateReq, err := http.NewRequest("POST",
		fmt.Sprintf("%s/panel/inbound/update/%d", m.serverURL, targetInbound.ID),
		bytes.NewBuffer(updateData))
	if err != nil {
		return "", fmt.Errorf("failed to create update request: %v", err)
	}

	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	updateResp, err := m.client.Do(updateReq)
	if err != nil {
		return "", fmt.Errorf("failed to send update request: %v", err)
	}
	defer updateResp.Body.Close()

	// Генерируем ссылку для подключения
	link := generateVmessLink(newClient.ID, email, targetInbound.Port)
	return link, nil
}

// Вспомогательные функции
func generateUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func generateVmessLink(clientID, email string, port int) string {
	config := VmessConfig{
		Version: "2",
		Name:    email,
		Address: "116.203.117.243",
		Port:    port,
		ID:      clientID,
		Aid:     0,
		Net:     "tcp",
		Type:    "none",
		Host:    "",
		Path:    "",
		TLS:     "none",
		SNI:     "",
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return ""
	}

	return "vmess://" + base64.StdEncoding.EncodeToString(configJSON)
}

// Обновляем метод GetClientInfo
func (m *XrayManager) GetClientInfo(email string) (*ClientInfo, error) {
	// Проверяем логин перед выполнением операции
	if err := m.checkAndRelogin(); err != nil {
		return nil, fmt.Errorf("failed to check login status: %v", err)
	}

	// Получаем список inbounds
	listReq, err := http.NewRequest("GET", m.serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := m.client.Do(listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp struct {
		Success bool      `json:"success"`
		Obj     []Inbound `json:"obj"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Ищем клиента во всех inbounds
	for _, inbound := range inboundResp.Obj {
		var settings VlessSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		for _, client := range settings.Clients {
			if client.Email == email {
				// Получаем статистику клиента
				statsReq, err := http.NewRequest("GET",
					fmt.Sprintf("%s/panel/api/inbounds/getClientTraffics/%s", m.serverURL, client.Email),
					nil)
				if err != nil {
					return nil, fmt.Errorf("failed to create stats request: %v", err)
				}

				statsReq.Header.Set("Content-Type", "application/json")
				statsReq.Header.Set("X-Requested-With", "XMLHttpRequest")

				statsResp, err := m.client.Do(statsReq)
				if err != nil {
					return nil, fmt.Errorf("failed to get client stats: %v", err)
				}
				defer statsResp.Body.Close()

				var statsData struct {
					Success bool `json:"success"`
					Obj     struct {
						Up   int64 `json:"up"`
						Down int64 `json:"down"`
					} `json:"obj"`
				}

				if err := json.NewDecoder(statsResp.Body).Decode(&statsData); err != nil {
					return nil, fmt.Errorf("failed to decode stats response: %v", err)
				}

				// Преобразуем время из миллисекунд в time.Time
				expiryTime := time.UnixMilli(client.ExpiryTime)
				if client.ExpiryTime == 0 {
					expiryTime = time.Now().AddDate(100, 0, 0) // 100 лет
				}

				return &ClientInfo{
					ID:         client.ID,
					Email:      client.Email,
					ExpiryTime: expiryTime,
					Enable:     client.Enable,
					Port:       inbound.Port,
					Up:         statsData.Obj.Up,
					Down:       statsData.Obj.Down,
					Total:      int64(client.TotalGB) * 1024 * 1024 * 1024,
					LimitIP:    client.LimitIP,
					TotalGB:    int64(client.TotalGB),
					Flow:       client.Flow,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("client with email %s not found", email)
}

// Добавляем метод для проверки и обновления логина
func (m *XrayManager) checkAndRelogin() error {
	// Пробуем сделать тестовый запрос
	listReq, err := http.NewRequest("GET", m.serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := m.client.Do(listReq)
	if err != nil || resp.StatusCode == http.StatusUnauthorized {
		// Если получили ошибку авторизации, пробуем залогиниться заново
		if err := m.login(); err != nil {
			return fmt.Errorf("failed to relogin: %v", err)
		}
	}

	return nil
}

// Обновляем метод, добавляя userID
func (m *XrayManager) GenerateVmessLink(uuid string, userID string, port int) string {
	config := map[string]interface{}{
		"v":    "2",
		"ps":   fmt.Sprintf("Proxy-%s", userID), // Используем строковый userID
		"add":  m.serverURL,
		"port": port,
		"id":   uuid,
		"aid":  "0",
		"net":  "ws",
		"type": "none",
		"host": m.serverURL,
		"path": "/",
		"tls":  "tls",
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		log.Printf("Error marshaling config: %v", err)
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return "vmess://" + encoded
}
