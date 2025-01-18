package xray

// XrayConfig представляет конфигурацию Xray сервера
type XrayConfig struct {
    Server    string     `json:"server"`    // Адрес сервера
    Inbounds  []Inbound  `json:"inbounds"`
    Outbounds []Outbound `json:"outbounds"`
}

// Inbound конфигурация входящего соединения
type Inbound struct {
    Port     int      `json:"port"`
    Protocol string   `json:"protocol"`
    Settings Settings `json:"settings"`
}

// Settings настройки для входящего соединения
type Settings struct {
    Clients []Client `json:"clients"`
}

// Client представляет конфигурацию клиента
type Client struct {
    ID      string `json:"id"`
    Email   string `json:"email"`
    Level   int    `json:"level"`
    AlterId int    `json:"alterId"`
}

// Outbound конфигурация исходящего соединения
type Outbound struct {
    Protocol string `json:"protocol"`
    Settings struct {
        Vnext []interface{} `json:"vnext"`
    } `json:"settings"`
} 