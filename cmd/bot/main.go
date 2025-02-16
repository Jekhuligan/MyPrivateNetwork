package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"myprivatenetwork/cmd/xray"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Добавим структуру для хранения состояния пользователя
type UserState struct {
	WaitingForStars bool
	StarsRequired   int
}

// Карта для хранения состояний пользователей
var userStates = make(map[int64]*UserState)

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
		if update.Message != nil {
			// Инициализируем состояние пользователя при каждом сообщении, если его нет
			userID := update.Message.Chat.ID
			state, exists := userStates[userID]
			if !exists {
				state = &UserState{
					WaitingForStars: false,
					StarsRequired:   5,
				}
				userStates[userID] = state
			}

			// Проверяем, есть ли в сообщении звезды
			if update.Message.Text != "" && strings.Contains(update.Message.Text, "⭐️") {
				stars := countStars(update.Message.Text)

				// Проверяем, ожидаем ли мы звезды от этого пользователя
				if state.WaitingForStars {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

					if stars >= state.StarsRequired {
						// Создаем подключение
						userEmail := fmt.Sprintf("tg_%d", update.Message.Chat.ID)
						link, err := xrayManager.CreateClient(userEmail)
						if err != nil {
							msg.Text = fmt.Sprintf("Ошибка при создании подключения: %v", err)
						} else {
							// Создаем QR код
							qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s",
								url.QueryEscape(link))

							// Отправляем фото с QR кодом
							photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileURL(qrURL))
							photo.Caption = fmt.Sprintf("Подключение создано успешно!\n\nКонфигурация:\n%s", link)
							bot.Send(photo)

							// Сбрасываем состояние ожидания
							state.WaitingForStars = false
						}
					} else {
						msg.Text = fmt.Sprintf("Недостаточно звезд. Необходимо %d ⭐️, получено %d ⭐️",
							state.StarsRequired, stars)
					}

					if _, err := bot.Send(msg); err != nil {
						log.Printf("Error sending message: %v", err)
					}
				}
			}

			// Обработка команд
			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				switch update.Message.Command() {
				case "start":
					msg.Text = "Привет! Я бот для управления VPN.\n" +
						"Доступные команды:\n" +
						"/create - создать подключение (стоимость: 5 ⭐️)\n" +
						"/qr - показать QR код для подключения\n" +
						"/info - информация о подключении\n\n" +
						"Для оплаты просто отправьте нужное количество звезд ⭐️"

				case "create":
					// Проверяем, есть ли уже активное подключение
					userEmail := fmt.Sprintf("tg_%d", update.Message.Chat.ID)
					exists, err := xrayManager.ClientExists(userEmail)
					if err != nil {
						msg.Text = fmt.Sprintf("Ошибка при проверке подключения: %v", err)
						break
					}

					if exists {
						msg.Text = "У вас уже есть активное подключение. Используйте /qr для получения данных подключения"
						break
					}

					// Запрашиваем оплату звездами
					state.WaitingForStars = true
					msg.Text = fmt.Sprintf("Для создания подключения отправьте %d ⭐️ в одном сообщении", state.StarsRequired)

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

					// Получаем информацию о существующем клиенте
					clientInfo, err := xrayManager.GetClientInfo(userID)
					if err != nil {
						msg.Text = "У вас нет активного подключения. Используйте /create для создания нового подключения"
						break
					}

					// Генерируем ссылку для существующего клиента
					link := xrayManager.GenerateVmessLink(clientInfo.ID, userID, clientInfo.Port)
					if link == "" {
						msg.Text = "Ошибка при генерации ссылки подключения"
						break
					}

					// Создаем QR код
					qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s",
						url.QueryEscape(link))

					// Отправляем фото
					photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileURL(qrURL))
					photo.Caption = fmt.Sprintf("Ваш QR код для подключения\nКонфигурация: %s", link)
					bot.Send(photo)

				case "info":
					userID := fmt.Sprintf("tg_%d", update.Message.Chat.ID)
					info, err := xrayManager.GetClientInfo(userID)
					if err != nil {
						msg.Text = fmt.Sprintf("Ошибка при получении информации: %v", err)
						break
					}

					// Форматируем размеры в удобочитаемый вид
					formatBytes := func(bytes int64) string {
						const unit = 1024
						if bytes < unit {
							return fmt.Sprintf("%d B", bytes)
						}
						div, exp := int64(unit), 0
						for n := bytes / unit; n >= unit; n /= unit {
							div *= unit
							exp++
						}
						return fmt.Sprintf("%.1f %cB",
							float64(bytes)/float64(div), "KMGTPE"[exp])
					}

					// Форматируем оставшееся время
					timeLeft := time.Until(info.ExpiryTime)
					daysLeft := int(timeLeft.Hours() / 24)

					msg.Text = fmt.Sprintf("Информация о подключении:\n\n"+
						"📊 Трафик:\n"+
						"↑ Отправлено: %s\n"+
						"↓ Получено: %s\n"+
						"💾 Общий лимит: %s\n\n"+
						"⌛️ Осталось дней: %d\n"+
						"�� Действует до: %s",
						formatBytes(info.Up),
						formatBytes(info.Down),
						formatBytes(info.Total),
						daysLeft,
						info.ExpiryTime.Format("02.01.2006 15:04"))

				default:
					msg.Text = "Неизвестная команда"
				}

				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			}
		}
	}
}

// Функция для подсчета звезд в сообщении
func countStars(text string) int {
	return strings.Count(text, "⭐")
}
