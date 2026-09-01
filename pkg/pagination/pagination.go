package pagination

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	maxbotModel "github.com/max-messenger/max-bot-api-client-go/v2/model"
)

// Paginator - основной объект пагинации
type Paginator struct {
	api               *maxbot.Api
	handler           PaginationHandler
	CenterButCallBack string
}

// NewPaginator - создание нового пагинатора
func NewPaginator(api *maxbot.Api, handler PaginationHandler) *Paginator {
	return &Paginator{
		api:     api,
		handler: handler,
	}
}

// WithCenterButton - добавляет центральную кнопку
func (p *Paginator) WithCenterButton(callback string) *Paginator {
	p.CenterButCallBack = callback
	return p
}

// SendPage - отправить страницу
func (p *Paginator) SendPage(ctx context.Context, chatID int64, userID int64, num int) (string, error) {
	page, err := p.handler.GetPage(ctx, userID, num)
	if err != nil {
		log.Printf("Ошибка! Для пользователя (%d) в чате(%d) из пагинатора не удалось получить страницу (%d)",
			userID, chatID, num)
		return "", err
	}

	text := p.buildMessage(page)

	keyboard := p.buildKeyboard(userID, num, page)

	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText(text).
		SetFormat(maxbotModel.FormatMarkdown).
		SetDisableLinkPreview(true)

	if keyboard != nil {
		msg.AddKeyboard(keyboard)
	}

	result, err := p.api.Messages.Send(ctx, msg)
	if err != nil {
		log.Printf("Ошибка! Для пользователя (%d) в чате(%d) из пагинатора не удалось отправить сообщение.",
			userID, chatID)
		return "", err
	}

	return result.Message.Body.Mid, nil
}

// UpdatePage - обновить существующую страницу
func (p *Paginator) UpdatePage(ctx context.Context, chatID int64, userID int64, messageID string, num int) error {
	page, err := p.handler.GetPage(ctx, userID, num)
	if err != nil {
		log.Printf("Ошибка! Для пользователя (%d) в чате(%d) из пагинатора не удалось получить страницу (%d)",
			userID, chatID, num)
		return err
	}

	text := p.buildMessage(page)

	keyboard := p.buildKeyboard(userID, num, page)

	body := maxbotModel.NewMessageBody{
		Text:        text,
		Format:      maxbotModel.FormatMarkdown,
		Attachments: []maxbotModel.Attachment{},
	}

	if keyboard != nil {
		body.Attachments = append(body.Attachments, keyboard.Build())
	}

	_, err = p.api.Messages.EditMessage(ctx, messageID, body)

	if err != nil {
		log.Printf("Ошибка! Для пользователя (%d) в чате(%d) из пагинатора не удалось обновить сообщение.",
			userID, chatID)
	}

	return err
}

// buildMessage - сборка сообщения
func (p *Paginator) buildMessage(page *Page) string {
	var sb strings.Builder

	sb.WriteString(p.handler.GetTextHeader(page))
	sb.WriteString("\n\n")

	for i, item := range page.PerPageItems {
		sb.WriteString(p.handler.GetTextItem(item, i))
		sb.WriteString("\n\n\n")
	}

	sb.WriteString(p.handler.GetTextFooter(page))

	return sb.String()
}

// buildKeyboard - сборка клавиатуры
func (p *Paginator) buildKeyboard(userID int64, pageNum int, page *Page) *maxbotModel.Keyboard {
	if !page.HasPrevPage && !page.HasNextPage {
		return nil
	}

	prb := page.PrevButSym
	if prb == "" {
		prb = "<"
	}

	nxb := page.NextButSym
	if nxb == "" {
		nxb = ">"
	}

	keyboard := maxbotModel.NewKeyboard()
	row := keyboard.AddRow()

	if page.HasPrevPage {
		row.AddCallBack(prb, fmt.Sprintf("page:%d:%d", userID, pageNum-1))
	}

	ctb := page.CentButSym
	if ctb != "" {
		row.AddCallBack(ctb, page.CentButCallback)
	}

	if page.HasNextPage {
		row.AddCallBack(nxb, fmt.Sprintf("page:%d:%d", userID, pageNum+1))
	}

	return keyboard
}

// HandleCallback - обработка callback пагинации
func (p *Paginator) HandleCallback(ctx context.Context, chatID int64, userID int64, messageID string, callbackData string) (bool, int, error) {
	if !strings.HasPrefix(callbackData, "page:") {
		return false, 0, nil
	}

	parts := strings.Split(callbackData, ":")
	if len(parts) != 3 {
		return false, 0, fmt.Errorf("invalid page callback: %s", callbackData)
	}

	callbackUserID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, 0, err
	}

	if callbackUserID != userID {
		return false, 0, fmt.Errorf("user mismatch")
	}

	page, err := strconv.Atoi(parts[2])
	if err != nil {
		return false, 0, err
	}

	if err := p.UpdatePage(ctx, chatID, userID, messageID, page); err != nil {
		return false, page, err
	}

	return true, page, nil
}
