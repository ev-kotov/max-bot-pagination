package pagination

import "context"

// Page - данные страницы
type Page struct {
	Num             int    // Номер страницы
	TotalItemsCount int    // Всего элементов
	TotalPagesCount int    // Всего страниц
	HasPrevPage     bool   // Есть ли предыдущая страница
	HasNextPage     bool   // Есть ли следующая страница
	PrevButSym      string // Символ для кнопки "Назад"
	NextButSym      string // Символ для кнопки "Вперёд"
	CentButSym      string // Символ центральной кноки
	CentButCallback string // Callback для центральной кнопки
	PerPageItems    []any  // Элементы на странице
}

// PaginationHandler - интерфейс для обработки страниц
type PaginationHandler interface {
	GetPerPageItemsCount() int
	GetPage(ctx context.Context, userID int64, num int) (*Page, error)
	GetTextItem(item any, index int) string
	GetTextHeader(page *Page) string
	GetTextFooter(page *Page) string
}
