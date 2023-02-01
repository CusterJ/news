package domain

type Data struct {
	Data Content `json:"data"`
}
type Content struct {
	Content ArticleParse `json:"content"`
}

type ArticleParse struct {
	Id          string      `json:"id"`
	URL         string      `json:"url"`
	Title       Title       `json:"title"`
	Description Description `json:"description"`
	Dates       DatesStr    `json:"dates"`
}

type DatesStr struct {
	Posted string `json:"posted"`
}
