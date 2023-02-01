package domain

type ArticlesResponse struct {
	Data    []Article
	Message string
	Count   int64
}
