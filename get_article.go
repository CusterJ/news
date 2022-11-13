package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Marshal API writer Article struct
type ArticleRespons struct {
	Message string  `json:"message,omitempty"`
	Data    Article `json:"data,omitempty"`
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
	data := Article{}
	// fmt.Println(string(body))
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println("Unmarshaling Article error ", err)
		return
	}
	Arts = append(Arts, data)
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
	// fmt.Println(Arts)
}
