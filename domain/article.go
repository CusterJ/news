package domain

type Article struct {
	Id          string      `json:"id" validate:"required,uuid4"`
	URL         string      `json:"url"` // validate if uri?
	Title       Title       `json:"title"`
	Description Description `json:"description"`
	Dates       Dates       `json:"dates"`
}
type Title struct {
	Short string `json:"short" validate:"required,min=3,max=400"`
}
type Description struct {
	Long string `json:"long" validate:"required,min=10"`
}
type Dates struct {
	Posted int64 `json:"posted" validate:"required,gt=1577840461"` // Date and time (GMT): Wednesday, January 1, 2020 1:01:01 AM
}
