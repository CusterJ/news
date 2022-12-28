package domain

// Marshal API writer Article struct
type ArticleResponse struct {
	Message string  `json:"message,omitempty"`
	Data    Article `json:"data,omitempty"`
}

type ArticlesResponse struct {
	Data  []Article
	Count int64
}

type ArticlesRequest struct {
	Skip  int
	Limit int
}

// Point Respons json
type Article struct {
	Data ArticleData `json:"data"`
}
type ArticleData struct {
	Content Content `json:"content"`
}
type Content struct {
	Id          string       `json:"id"`
	URL         string       `json:"url"`
	Title       ArticleTitle `json:"title"`
	Description Description  `json:"description"`
}
type ArticleTitle struct {
	Short string `json:"short"`
}
type Description struct {
	Long string `json:"long"`
}
