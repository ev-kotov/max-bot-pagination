package pagination

import (
	"context"
	"testing"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	maxbotModel "github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Моки ---

type MockMessagesAPI struct {
	mock.Mock
}

func (m *MockMessagesAPI) GetMessages(ctx context.Context, chatID, from, to, count int64, messageIDs []string) (maxbotModel.MessageList, error) {
	args := m.Called(ctx, chatID, from, to, count, messageIDs)
	return args.Get(0).(maxbotModel.MessageList), args.Error(1)
}

func (m *MockMessagesAPI) GetMessageByID(ctx context.Context, messageID string) (maxbotModel.Message, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(maxbotModel.Message), args.Error(1)
}

func (m *MockMessagesAPI) Send(ctx context.Context, msg *maxbot.Message) (maxbotModel.SendMessageResult, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(maxbotModel.SendMessageResult), args.Error(1)
}

func (m *MockMessagesAPI) EditMessage(ctx context.Context, messageID string, body maxbotModel.NewMessageBody) (maxbotModel.SimpleQueryResult, error) {
	args := m.Called(ctx, messageID, body)
	return args.Get(0).(maxbotModel.SimpleQueryResult), args.Error(1)
}

func (m *MockMessagesAPI) DeleteMessage(ctx context.Context, messageID string) (maxbotModel.SimpleQueryResult, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(maxbotModel.SimpleQueryResult), args.Error(1)
}

func (m *MockMessagesAPI) AnswerOnCallback(ctx context.Context, callbackID string, answer maxbotModel.CallbackAnswer) (maxbotModel.SimpleQueryResult, error) {
	args := m.Called(ctx, callbackID, answer)
	return args.Get(0).(maxbotModel.SimpleQueryResult), args.Error(1)
}

func (m *MockMessagesAPI) GetVideoAttachmentDetails(ctx context.Context, videoToken string) (maxbotModel.VideoAttachmentDetails, error) {
	args := m.Called(ctx, videoToken)
	return args.Get(0).(maxbotModel.VideoAttachmentDetails), args.Error(1)
}

type MockPaginationHandler struct {
	mock.Mock
}

func (m *MockPaginationHandler) GetPerPageItemsCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaginationHandler) GetPage(ctx context.Context, userID int64, num int) (*Page, error) {
	args := m.Called(ctx, userID, num)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPaginationHandler) GetTextItem(item any, index int) string {
	args := m.Called(item, index)
	return args.String(0)
}

func (m *MockPaginationHandler) GetTextHeader(page *Page) string {
	args := m.Called(page)
	return args.String(0)
}

func (m *MockPaginationHandler) GetTextFooter(page *Page) string {
	args := m.Called(page)
	return args.String(0)
}

// --- Тесты ---

func TestBuildMessage(t *testing.T) {
	mockHandler := new(MockPaginationHandler)
	paginator := NewPaginator(&maxbot.Api{}, mockHandler)

	page := &Page{Num: 1, PerPageItems: []any{"item1", "item2"}}

	mockHandler.On("GetTextHeader", page).Return("Header")
	mockHandler.On("GetTextItem", "item1", 0).Return("Item 1")
	mockHandler.On("GetTextItem", "item2", 1).Return("Item 2")
	mockHandler.On("GetTextFooter", page).Return("Footer")

	result := paginator.buildMessage(page)

	assert.Contains(t, result, "Header")
	assert.Contains(t, result, "Item 1")
	assert.Contains(t, result, "Item 2")
	assert.Contains(t, result, "Footer")
	mockHandler.AssertExpectations(t)
}

func TestBuildKeyboard(t *testing.T) {
	mockHandler := new(MockPaginationHandler)
	paginator := NewPaginator(&maxbot.Api{}, mockHandler)

	t.Run("with prev and next", func(t *testing.T) {
		page := &Page{Num: 2, HasPrevPage: true, HasNextPage: true, PrevButSym: "<<", NextButSym: ">>"}
		assert.NotNil(t, paginator.buildKeyboard(123, 2, page))
	})

	t.Run("no prev and no next", func(t *testing.T) {
		page := &Page{Num: 1, HasPrevPage: false, HasNextPage: false}
		assert.Nil(t, paginator.buildKeyboard(123, 1, page))
	})

	t.Run("with center button", func(t *testing.T) {
		page := &Page{Num: 1, HasPrevPage: false, HasNextPage: true, CentButSym: "Center", CentButCallback: "center_cb"}
		assert.NotNil(t, paginator.buildKeyboard(123, 1, page))
	})
}

func TestHandleCallback(t *testing.T) {
	mockHandler := new(MockPaginationHandler)
	mockMessages := new(MockMessagesAPI)
	paginator := NewPaginator(&maxbot.Api{Messages: mockMessages}, mockHandler)

	ctx := context.Background()
	chatID, userID, messageID := int64(1), int64(100), "msg123"

	t.Run("invalid prefix", func(t *testing.T) {
		handled, page, err := paginator.HandleCallback(ctx, chatID, userID, messageID, "other:data")
		assert.False(t, handled)
		assert.Equal(t, 0, page)
		assert.NoError(t, err)
	})

	t.Run("user mismatch", func(t *testing.T) {
		handled, page, err := paginator.HandleCallback(ctx, chatID, userID, messageID, "page:200:2")
		assert.False(t, handled)
		assert.Equal(t, 0, page)
		assert.ErrorContains(t, err, "user mismatch")
	})

	t.Run("successful update", func(t *testing.T) {
		pageData := &Page{Num: 2, HasPrevPage: true, HasNextPage: false}
		mockHandler.On("GetPage", ctx, userID, 2).Return(pageData, nil)
		mockHandler.On("GetTextHeader", pageData).Return("Header")
		mockHandler.On("GetTextFooter", pageData).Return("Footer")
		mockMessages.On("EditMessage", ctx, messageID, mock.Anything).Return(maxbotModel.SimpleQueryResult{Success: true}, nil)

		handled, page, err := paginator.HandleCallback(ctx, chatID, userID, messageID, "page:100:2")

		assert.True(t, handled)
		assert.Equal(t, 2, page)
		assert.NoError(t, err)
		mockHandler.AssertExpectations(t)
		mockMessages.AssertExpectations(t)
	})
}

func TestSendPage(t *testing.T) {
	mockHandler := new(MockPaginationHandler)
	mockMessages := new(MockMessagesAPI)
	paginator := NewPaginator(&maxbot.Api{Messages: mockMessages}, mockHandler)

	ctx := context.Background()
	pageData := &Page{Num: 1, PerPageItems: []any{"item1"}}

	mockHandler.On("GetPage", ctx, int64(100), 1).Return(pageData, nil)
	mockHandler.On("GetTextHeader", pageData).Return("Header")
	mockHandler.On("GetTextItem", "item1", 0).Return("Item 1")
	mockHandler.On("GetTextFooter", pageData).Return("Footer")
	mockMessages.On("Send", ctx, mock.Anything).Return(maxbotModel.SendMessageResult{
		Message: maxbotModel.Message{Body: maxbotModel.MessageBody{Mid: "msg123"}},
	}, nil)

	msgID, err := paginator.SendPage(ctx, 1, 100, 1)

	assert.NoError(t, err)
	assert.Equal(t, "msg123", msgID)
	mockHandler.AssertExpectations(t)
	mockMessages.AssertExpectations(t)
}
