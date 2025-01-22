package main

import (
	"fmt"
	"log"
	"myprivatenetwork/internal/xray"
	"net/url"
)

func main() {
	// Создаем менеджер
	manager, err := xray.NewXrayManager(SERVER_URL, USERNAME, PASSWORD)
	if err != nil {
		log.Fatalf("Failed to create XrayManager: %v", err)
	}

	// Создаем нового клиента (теперь с автоматической проверкой и удалением)
	link, err := manager.CreateClient("golang")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Выводим результаты
	fmt.Printf("\nClient connection details:\n")
	fmt.Printf("Link: %s\n", link)
	fmt.Printf("QR code: https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s\n",
		url.QueryEscape(link))
}
