package domain

type Article struct {
	Id          string      `json:"id"`
	URL         string      `json:"url"`
	Title       Title       `json:"title"`
	Description Description `json:"description"`
	Dates       Dates       `json:"dates"`
}
type Title struct {
	Short string `json:"short"`
}
type Description struct {
	Long string `json:"long"`
}
type Dates struct {
	Posted int64 `json:"posted"`
}
