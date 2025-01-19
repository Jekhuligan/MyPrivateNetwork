package main

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
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

type CreateUserRequest struct {
    Up        int64  `json:"up"`        // лимит загрузки
    Down      int64  `json:"down"`      // лимит скачивания
    Total     int64  `json:"total"`     // общий трафик
    Remark    string `json:"remark"`    // имя пользователя
    Enable    bool   `json:"enable"`    // активен ли пользователь
    ExpiryTime int64  `json:"expiryTime"` // время истечения
    Listen    string `json:"listen"`    // адрес прослушивания
    Port      int    `json:"port"`      // порт
    Protocol  string `json:"protocol"`  // протокол
    Settings  string `json:"settings"`  // настройки в JSON
}

func main() {
    // Создаем HTTP клиент
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{
        Transport: tr,
        Timeout:   time.Second * 10,
    }

    // 1. Выполняем логин
    loginReq := LoginRequest{
        Username: USERNAME,
        Password: PASSWORD,
    }

    loginData, err := json.Marshal(loginReq)
    if err != nil {
        log.Fatalf("Failed to marshal login request: %v", err)
    }

    loginResp, err := client.Post(SERVER_URL+"/login", "application/json", bytes.NewBuffer(loginData))
    if err != nil {
        log.Fatalf("Failed to send login request: %v", err)
    }
    defer loginResp.Body.Close()

    fmt.Printf("Login Status: %s\n", loginResp.Status)
    loginBody, _ := io.ReadAll(loginResp.Body)
    fmt.Printf("Login Response: %s\n\n", string(loginBody))

    if loginResp.StatusCode != http.StatusOK {
        log.Fatalf("Login failed")
    }

    // Сохраняем куки
    cookies := loginResp.Cookies()

    // 2. Получаем информацию о панели
    fmt.Println("Getting panel info...")
    panelReq, _ := http.NewRequest("GET", SERVER_URL+"/panel/", nil)
    for _, cookie := range cookies {
        panelReq.AddCookie(cookie)
    }

    panelResp, err := client.Do(panelReq)
    if err != nil {
        log.Printf("Failed to get panel info: %v", err)
    } else {
        panelBody, _ := io.ReadAll(panelResp.Body)
        panelResp.Body.Close()
        fmt.Printf("Panel Status: %s\n", panelResp.Status)
        fmt.Printf("Panel Response length: %d bytes\n\n", len(panelBody))
    }

    // 3. Создаем пользователя
    createReq := CreateUserRequest{
        Up:         0,           // без ограничений
        Down:       0,           // без ограничений
        Total:      0,           // без ограничений
        Remark:     "User@net.com",
        Enable:     true,
        ExpiryTime: time.Now().AddDate(0, 1, 0).Unix() * 1000, // +1 месяц
        Protocol:   "vmess",
        Port:       30249,
    }

    createData, err := json.Marshal(createReq)
    if err != nil {
        log.Fatalf("Failed to marshal create user request: %v", err)
    }

    // Пробуем создать пользователя
    fmt.Println("Creating user...")
    createReqObj, err := http.NewRequest("POST", SERVER_URL+"/panel/inbound/add", bytes.NewBuffer(createData))
    if err != nil {
        log.Fatalf("Failed to create request: %v", err)
    }

    // Добавляем куки и заголовки
    for _, cookie := range cookies {
        createReqObj.AddCookie(cookie)
    }
    createReqObj.Header.Set("Content-Type", "application/json")
    createReqObj.Header.Set("X-Requested-With", "XMLHttpRequest")

    createResp, err := client.Do(createReqObj)
    if err != nil {
        log.Fatalf("Failed to send create request: %v", err)
    }
    defer createResp.Body.Close()

    createBody, _ := io.ReadAll(createResp.Body)
    fmt.Printf("Create Status: %s\n", createResp.Status)
    fmt.Printf("Create Response: %s\n", string(createBody))
} 