package pagination

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
)

// API представляет "фасад" для работы с пагинацией (цепочечный (chainable) интерфейс).
type API struct {
	paginator *Paginator
}

// NewAPI создает новый экземпляр API пагинации.
func NewAPI(api *maxbot.Api, handler PaginationHandler) *API {
	return &API{
		paginator: NewPaginator(api, handler),
	}
}

// WithCenterButton добавляет callback для центральной кнопки и возвращает API для цепочки вызовов.
func (a *API) WithCenterButton(callback string) *API {
	a.paginator.WithCenterButton(callback)
	return a
}

// Send отправляет новое сообщение с пагинацией.
func (a *API) Send(ctx context.Context, chatID int64, userID int64, pageNum int) (string, error) {
	return a.paginator.SendPage(ctx, chatID, userID, pageNum)
}

// Edit обновляет существующее сообщение с пагинацией.
func (a *API) Edit(ctx context.Context, chatID int64, userID int64, messageID string, pageNum int) error {
	return a.paginator.UpdatePage(ctx, chatID, userID, messageID, pageNum)
}

// Handle обрабатывает callback-запрос пагинации.
// Возвращает:
//   - handled: был ли это callback-запрос именно этой пагинации
//   - pageNum: номер запрошенной страницы
//   - err: ошибка выполнения (если handled == true)
func (a *API) Handle(ctx context.Context, chatID int64, userID int64,
	messageID string, callbackData string) (bool, int, error) {
	return a.paginator.HandleCallback(ctx, chatID, userID, messageID, callbackData)
}
