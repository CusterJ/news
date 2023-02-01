package domain

type ArticleResponse struct {
	Message string  `json:"message,omitempty"`
	Data    Article `json:"data,omitempty"`
}
