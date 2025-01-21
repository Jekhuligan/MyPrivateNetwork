package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const (
	SERVER_URL = "https://116.203.117.243:30249/insurmountablemountain"
	USERNAME   = "tay"
	PASSWORD   = "tibmox-qAdmog-1xytba"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool    `json:"success"`
	Msg     string  `json:"msg"`
	Obj     *string `json:"obj"`
}

type InboundResponse struct {
	Success bool      `json:"success"`
	Msg     string    `json:"msg"`
	Obj     []Inbound `json:"obj"`
}

type Inbound struct {
	ID             int    `json:"id"`
	Up             int64  `json:"up"`
	Down           int64  `json:"down"`
	Total          int64  `json:"total"`
	Remark         string `json:"remark"`
	Enable         bool   `json:"enable"`
	ExpiryTime     int64  `json:"expiryTime"`
	Listen         string `json:"listen"`
	Port           int    `json:"port"`
	Protocol       string `json:"protocol"`
	Settings       string `json:"settings"`
	StreamSettings string `json:"streamSettings"`
	Tag            string `json:"tag"`
	Sniffing       string `json:"sniffing"`
}

type Client struct {
	ID      string `json:"id"`
	AlterId int    `json:"alterId"`
	Email   string `json:"email"`
}

type Settings struct {
	Clients []Client `json:"clients"`
}

type VlessSettings struct {
	Clients    []VlessClient `json:"clients"`
	Decryption string        `json:"decryption"`
	Fallbacks  []interface{} `json:"fallbacks"`
}

type VlessClient struct {
	ID         string `json:"id"`
	Flow       string `json:"flow"`
	Email      string `json:"email"`
	LimitIP    int    `json:"limitIp"`
	TotalGB    int    `json:"totalGB"`
	ExpiryTime int64  `json:"expiryTime"`
	Enable     bool   `json:"enable"`
	TgID       string `json:"tgId"`
	SubID      string `json:"subId"`
	Reset      int    `json:"reset"`
}

type VmessConfig struct {
	Version string `json:"v"`
	Name    string `json:"ps"`
	Address string `json:"add"`
	Port    int    `json:"port"`
	ID      string `json:"id"`
	Aid     int    `json:"aid"`
	Net     string `json:"net"`
	Type    string `json:"type"`
	Host    string `json:"host"`
	Path    string `json:"path"`
	TLS     string `json:"tls"`
	SNI     string `json:"sni"`
}

func generateUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func getClientQR(client *http.Client, serverURL, email string) (string, error) {
	// Получаем список inbounds
	listReq, err := http.NewRequest("GET", serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := client.Do(listReq)
	if err != nil {
		return "", fmt.Errorf("failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp struct {
		Success bool      `json:"success"`
		Obj     []Inbound `json:"obj"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Ищем клиента во всех inbounds
	for _, inbound := range inboundResp.Obj {
		var settings VlessSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		for _, client := range settings.Clients {
			if client.Email == email {
				// Создаем конфигурацию vmess
				config := VmessConfig{
					Version: "2",
					Name:    client.Email,
					Address: "116.203.117.243",
					Port:    inbound.Port,
					ID:      client.ID,
					Aid:     0,
					Net:     "tcp",
					Type:    "none",
					Host:    "",
					Path:    "",
					TLS:     "none",
					SNI:     "116.203.117.243",
				}

				// Преобразуем конфигурацию в JSON
				configJSON, err := json.Marshal(config)
				if err != nil {
					return "", fmt.Errorf("failed to marshal config: %v", err)
				}

				// Формируем ссылку vmess
				// vmess://base64(json-config)
				link := "vmess://" + base64.StdEncoding.EncodeToString(configJSON)

				return link, nil
			}
		}
	}

	return "", fmt.Errorf("client with email %s not found", email)
}

// Добавьте новую функцию для удаления клиентов
func deleteClientsByEmail(client *http.Client, serverURL, email string) error {
	// Получаем список inbounds
	listReq, err := http.NewRequest("GET", serverURL+"/panel/api/inbounds/list", nil)
	if err != nil {
		return fmt.Errorf("failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := client.Do(listReq)
	if err != nil {
		return fmt.Errorf("failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp struct {
		Success bool      `json:"success"`
		Obj     []Inbound `json:"obj"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Проходим по всем inbounds и ищем клиентов для удаления
	for _, inbound := range inboundResp.Obj {
		var settings VlessSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		// Ищем клиентов с указанным email
		var clientsUpdated bool
		var newClients []VlessClient
		for _, c := range settings.Clients {
			if c.Email != email {
				newClients = append(newClients, c)
			} else {
				clientsUpdated = true
				fmt.Printf("Found client to delete: %s in inbound %d\n", email, inbound.ID)
			}
		}

		// Если нашли и удалили клиентов, обновляем inbound
		if clientsUpdated {
			settings.Clients = newClients
			newSettings, err := json.Marshal(settings)
			if err != nil {
				return fmt.Errorf("failed to marshal new settings: %v", err)
			}
			inbound.Settings = string(newSettings)

			// Отправляем запрос на обновление
			updateData, err := json.Marshal(inbound)
			if err != nil {
				return fmt.Errorf("failed to marshal update request: %v", err)
			}

			updateReq, err := http.NewRequest("POST",
				fmt.Sprintf("%s/panel/inbound/update/%d", serverURL, inbound.ID),
				bytes.NewBuffer(updateData))
			if err != nil {
				return fmt.Errorf("failed to create update request: %v", err)
			}

			updateReq.Header.Set("Content-Type", "application/json")
			updateReq.Header.Set("X-Requested-With", "XMLHttpRequest")

			updateResp, err := client.Do(updateReq)
			if err != nil {
				return fmt.Errorf("failed to send update request: %v", err)
			}
			defer updateResp.Body.Close()

			updateBody, _ := io.ReadAll(updateResp.Body)
			fmt.Printf("Update Status for inbound %d: %s\n", inbound.ID, updateResp.Status)
			fmt.Printf("Update Response: %s\n", string(updateBody))
		}
	}

	return nil
}

// Добавим функцию для повторных попыток
func getClientQRWithRetry(client *http.Client, serverURL, email string, maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempt %d of %d to get QR code...\n", i+1, maxRetries)

		// Пауза перед повторной попыткой (увеличивается с каждой попыткой)
		if i > 0 {
			time.Sleep(time.Duration(i+2) * time.Second)
		}

		qr, err := getClientQR(client, serverURL, email)
		if err == nil {
			return qr, nil
		}
		lastErr = err
		fmt.Printf("Attempt %d failed: %v\n", i+1, err)
	}
	return "", fmt.Errorf("failed after %d attempts, last error: %v", maxRetries, lastErr)
}

// Упрощаем функцию создания QR кода - используем данные напрямую
func generateVmessLink(clientID string, email string, port int) string {
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
		TLS:     "tls",
		SNI:     "116.203.117.243",
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	return "vmess://" + base64.StdEncoding.EncodeToString(configJSON)
}

func main() {
	// Создаем HTTP клиент
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
		Jar:       jar,
	}

	// 1. Логин
	loginReq := LoginRequest{
		Username: USERNAME,
		Password: PASSWORD,
	}

	loginData, err := json.Marshal(loginReq)
	if err != nil {
		log.Fatalf("Failed to marshal login request: %v", err)
	}

	resp, err := client.Post(SERVER_URL+"/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		log.Fatalf("Failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	// 2. Получаем страницу панели для установки куки
	_, err = client.Get(SERVER_URL + "/panel")
	if err != nil {
		log.Fatalf("Failed to get panel page: %v", err)
	}

	// 3. Получаем список inbounds
	fmt.Println("Getting inbounds list...")
	listReq, err := http.NewRequest("GET", SERVER_URL+"/panel/api/inbounds/list", nil)
	if err != nil {
		log.Fatalf("Failed to create list request: %v", err)
	}

	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	listResp, err := client.Do(listReq)
	if err != nil {
		log.Fatalf("Failed to get inbounds list: %v", err)
	}
	defer listResp.Body.Close()

	var inboundResp struct {
		Success bool      `json:"success"`
		Msg     string    `json:"msg"`
		Obj     []Inbound `json:"obj"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&inboundResp); err != nil {
		log.Fatalf("Failed to decode inbounds response: %v", err)
	}

	// Вместо поиска по имени "TEST", ищем inbound с id=3
	var targetInbound *Inbound
	for i := range inboundResp.Obj {
		if inboundResp.Obj[i].ID == 3 {
			targetInbound = &inboundResp.Obj[i]
			break
		}
	}

	if targetInbound == nil {
		log.Fatalf("Inbound with ID=3 not found")
	}

	fmt.Printf("Found inbound: ID=%d, Remark=%s, Protocol=%s\n",
		targetInbound.ID,
		targetInbound.Remark,
		targetInbound.Protocol,
	)

	// Парсим текущие настройки
	var settings VlessSettings
	if err := json.Unmarshal([]byte(targetInbound.Settings), &settings); err != nil {
		log.Fatalf("Failed to parse settings: %v", err)
	}

	// Перед созданием нового клиента удаляем существующих с таким же email
	fmt.Println("Checking and deleting existing clients with email 'golang'...")
	if err := deleteClientsByEmail(client, SERVER_URL, "golang"); err != nil {
		log.Fatalf("Failed to delete existing clients: %v", err)
	}

	// Вычисляем дату истечения (1 месяц от текущей даты)
	expiryTime := time.Now().AddDate(0, 1, 0).UnixMilli()

	// Создаем нового клиента
	newClient := VlessClient{
		ID:         generateUUID(),
		Flow:       "xtls-rprx-vision",
		Email:      "golang",
		LimitIP:    0,
		TotalGB:    0,
		ExpiryTime: expiryTime,
		Enable:     true,
		TgID:       "",
		SubID:      "",
		Reset:      0,
	}

	settings.Clients = append(settings.Clients, newClient)

	// Обновляем настройки
	newSettings, err := json.Marshal(settings)
	if err != nil {
		log.Fatalf("Failed to marshal new settings: %v", err)
	}
	targetInbound.Settings = string(newSettings)

	// Отправляем запрос на обновление
	updateData, err := json.Marshal(targetInbound)
	if err != nil {
		log.Fatalf("Failed to marshal update request: %v", err)
	}

	fmt.Println("Updating inbound...")
	updateReq, err := http.NewRequest("POST", fmt.Sprintf("%s/panel/inbound/update/%d", SERVER_URL, targetInbound.ID), bytes.NewBuffer(updateData))
	if err != nil {
		log.Fatalf("Failed to create update request: %v", err)
	}

	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	updateResp, err := client.Do(updateReq)
	if err != nil {
		log.Fatalf("Failed to send update request: %v", err)
	}
	defer updateResp.Body.Close()

	updateBody, _ := io.ReadAll(updateResp.Body)
	fmt.Printf("Update Status: %s\n", updateResp.Status)
	fmt.Printf("Update Response: %s\n", string(updateBody))

	// Генерируем QR код сразу из данных нового клиента
	qrLink := generateVmessLink(newClient.ID, newClient.Email, targetInbound.Port)

	fmt.Printf("\nClient connection details:\n")
	fmt.Printf("Link: %s\n", qrLink)
	fmt.Printf("To get QR code, visit: https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s\n",
		url.QueryEscape(qrLink))
}
