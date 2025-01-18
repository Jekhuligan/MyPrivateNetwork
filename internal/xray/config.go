package xray

import (
    "encoding/json"
    "fmt"
    "os"
)

type XrayConfig struct {
    Address string `json:"address"`
    Inbounds  []Inbound  `json:"inbounds"`
    Outbounds []Outbound `json:"outbounds"`
}

func (c *XrayConfig) SaveToFile(path string) error {
    data, err := json.MarshalIndent(c, "", "    ")
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    return os.WriteFile(path, data, 0644)
}

type Inbound struct {
    Port     int      `json:"port"`
    Protocol string   `json:"protocol"`
    Settings Settings `json:"settings"`
}

type Settings struct {
    Clients []Client `json:"clients"`
}

type Client struct {
    ID      string `json:"id"`
    Email   string `json:"email"`
    Level   int    `json:"level"`
    AlterId int    `json:"alterId"`
}

type Outbound struct {
    Protocol string `json:"protocol"`
    Settings struct {
        Vnext []interface{} `json:"vnext"`
    } `json:"settings"`
} 