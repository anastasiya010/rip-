package main

// Модель услуги (лекарство)
type Service struct {
	ID           int
	Name         string
	ImageKey     string // ключ в Minio
	Category     string
	Manufacturer string
	ShortInfo    string
	Description  string
}

// Модель заявки (корзины)
type Order struct {
	ID         int
	Title      string
	Comment    string
	ResultNote string // поле результата вычислений (пока фиксированное)
	Items      []OrderItem
}

// Элемент заявки (плоская структура, без вложенных массивов в услуге)
type OrderItem struct {
	ServiceID   int
	ServiceName string
	ImageKey    string
	Quantity    int
	Note        string // поле м-м (комментарий/результат)
}
