package newsparser

func Start() {
	GetArticles(GetNewsList(NewsQuery(15, 0, 1666056885)))
}
