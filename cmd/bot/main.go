package main

import (
	"fmt"
	"log"
	"net/url"

	"myprivatenetwork/internal/xray"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Инициализируем XRay менеджер
	xrayManager, err := xray.NewXrayManager(SERVER_URL, USERNAME, PASSWORD)
	if err != nil {
		log.Fatalf("Failed to create XRay manager: %v", err)
	}

	// Инициализируем бота
	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Обрабатываем команды
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "start":
				msg.Text = "Привет! Я бот для управления VPN.\n" +
					"Доступные команды:\n" +
					"/create - создать подключение\n" +
					"/qr - показать QR код для подключения"

			case "create":
				// Используем ID чата как идентификатор пользователя
				userID := fmt.Sprintf("tg_%d", update.Message.Chat.ID)
				link, err := xrayManager.CreateClient(userID)
				if err != nil {
					msg.Text = fmt.Sprintf("Ошибка при создании подключения: %v", err)
					break
				}

				// Создаем QR код
				qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s",
					url.QueryEscape(link))

				// Отправляем фото с QR кодом
				photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileURL(qrURL))
				photo.Caption = "Подключение создано успешно!\nОтсканируйте QR код для подключения"
				bot.Send(photo)
				continue

			case "qr":
				userID := fmt.Sprintf("tg_%d", update.Message.Chat.ID)
				exists, err := xrayManager.ClientExists(userID)
				if err != nil {
					msg.Text = fmt.Sprintf("Ошибка при проверке подключения: %v", err)
					break
				}

				if !exists {
					msg.Text = "У вас нет активного подключения. Используйте /create для создания"
					break
				}

				// Получаем ссылку для клиента
				link, err := xrayManager.CreateClient(userID)
				if err != nil {
					msg.Text = fmt.Sprintf("Ошибка при получении данных подключения: %v", err)
					break
				}

				// Создаем QR код
				qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s",
					url.QueryEscape(link))

				// Отправляем фото
				photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileURL(qrURL))
				photo.Caption = fmt.Sprintf("Ваш QR код для подключения\nКонфигурация: %s", link)
				bot.Send(photo)
				continue

			default:
				msg.Text = "Неизвестная команда"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}
