# Библиотека пагинации для MAX Bot

Библиотека для реализации пагинации сообщений в чат-ботах на базе [MAX Bot API](https://github.com/max-messenger/max-bot-api-client-go).

## Возможности

- **Гибкая кастомизация**: полный контроль над заголовком, футером и отображением каждого элемента страницы через интерфейс `PaginationHandler`
- **Автоматическая клавиатура**: генерация inline-кнопок "Назад", "Вперёд" и центральной кнопки (опционально)
- **Бесшовное обновление**: метод `Edit` для мгновенного обновления сообщения при переключении страниц без отправки новых
- **Безопасность**: встроенная проверка `userID` в callback-запросах для предотвращения обработки чужих нажатий

## Установка

```bash
go get github.com/ev-kotov/max-bot-pagination/pkg/pagination max-bot-pagination
```

## Пример использования

Создайте структуру, которая реализует интерфейс `PaginationHandler` для ваших данных:

```go
package main

import (
 "context"
 "fmt"
 
 "github.com/your-username/your-repo/pkg/pagination"
 maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
)

type MyDataHandler struct {
 items []string // здесь Ваши данные
}

func (h *MyDataHandler) GetPerPageItemsCount() int {
 return 3 // Отобразим три сущности на странице
}

func (h *MyDataHandler) GetPage(ctx context.Context, userID int64, num int) (*pagination.Page, error) {
 start := (num - 1) * h.GetPerPageItemsCount()
 end := start + h.GetPerPageItemsCount()
 
 // ... ваша логика получения среза и проверки границ ...

 return &pagination.Page{
  Num:             num,
  TotalItemsCount: len(h.items),
  TotalPagesCount: 10,
  HasPrevPage:     num > 1,
  HasNextPage:     num < 10,
  PerPageItems:    []any{"Элемент А", "Элемент Б"}, 
 }, nil
}

func (h *MyDataHandler) GetTextItem(item any, index int) string {
 return "Здесь основной текст Вашей сущности"
}

func (h *MyDataHandler) GetTextHeader(page *pagination.Page) string {
 return "Здесь шапка"
}

func (h *MyDataHandler) GetTextFooter(page *pagination.Page) string {
 return "А здесь подвал"
}
```

Инициализация и отправка:

```go
api, err := maxbot.NewApi("YOUR_BOT_TOKEN")
if err != nil {
    log.Fatal(err)
}

// Создание обработчика данных
handler := &MyDataHandler{items: myItems}

// Создание пагинатора через удобный API-фасад
pager := pagination.NewAPI(api, handler).WithCenterButton("🏠 Главное меню")

// Отправка первой страницы
messageID, err := pager.Send(ctx, chatID, userID, 1)
if err != nil {
 log.Printf("Ошибка отправки: %v", err)
}
```

Обработка callback-запросов, в вашем обработчике обновлений:

```go
func handleCallback(ctx context.Context, update maxbotModel.Update) {
 userID := update.UserID
 chatID := update.ChatID
 messageID := update.CallbackQuery.Message.ID
 callbackData := update.CallbackQuery.Data

 // Передаем управление пагинатору
 handled, pageNum, err := pager.Handle(ctx, chatID, userID, messageID, callbackData)
 
 if handled {
  if err != nil {
   log.Printf("Ошибка пагинации: %v", err)
  }
  return // Пагинатор успешно перехватил и обновил сообщение
 }

 // Обработка других callback-запросов вашего бота...
}
```
