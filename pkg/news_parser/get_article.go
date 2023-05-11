package newsparser

import (
	"News/domain"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func getArticles(list []string) (news []domain.Article) {
	for _, id := range list {
		art, err := getArticle(id)
		if err != nil {
			log.Println("error get article from site: ", err)

			continue
		}

		news = append(news, art)
	}

	return
}

func getArticle(id string) (adb domain.Article, err error) {
	ArticlePayload := strings.NewReader(articleQuery(id))

	req, err := http.NewRequest(http.MethodPost, Url, ArticlePayload)
	if err != nil {
		fmt.Println("Wrap request error", err)

		return adb, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Requerst for article error: ", err)

		return adb, err
	}

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	// fmt.Println(string(body))

	data := domain.Data{}

	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println("Unmarshaling Article error ", err)
	}
	// fmt.Println("func GetArticle", data.Data.Content.Title.Short)

	adb.Id = data.Data.Content.Id
	adb.URL = data.Data.Content.URL
	adb.Title = data.Data.Content.Title
	adb.Description = data.Data.Content.Description

	// convert date from string to int64
	dateTS, err := strconv.Atoi(data.Data.Content.Dates.Posted)
	if err != nil {
		return
	}

	adb.Dates.Posted = int64(dateTS)

	return adb, err
}
