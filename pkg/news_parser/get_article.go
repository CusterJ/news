package newsparser

import (
	"News/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func GetArticles(list []string) {
	fmt.Println("func GetArticles from", list)
	for _, id := range list {
		GetArticle(id)
	}

}

func GetArticle(id string) {
	ArticlePayload := strings.NewReader(ArticleQuery(id))
	req, err := http.NewRequest("POST", Url, ArticlePayload)
	if err != nil {
		fmt.Println("Wrap request error", err)
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Requerst error, stop program")
		panic(err)
	}

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	data := domain.Article{}
	// fmt.Println(string(body))
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println("Unmarshaling Article error ", err)
		return
	}
	News = append(News, data)
	fmt.Println("func GetArticle", data.Data.Content.Title.Short)
	// GetArticleById("415695b0-0dec-4d01-b8d7-e2c7ba7700ce")
	// ok := FindAndInsert(data)
	// fmt.Println("FindAndInsert result: ", ok)
	// err = UpdateOne(data)
	// if err != nil {
	// 	fmt.Println("UpdateOne call err: ", err)
	// 	return
	// }
}
