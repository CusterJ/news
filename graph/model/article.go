package model

type Article struct {
	Id          string `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        int64  `json:"date"`
}
