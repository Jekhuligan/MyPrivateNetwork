package xray

// Здесь разместите все структуры из main.go:
// LoginRequest, LoginResponse, InboundResponse, Inbound,
// Client, Settings, VlessSettings, VlessClient, VmessConfig

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
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
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
