package main

import (
	"News/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Marshal API writer Newslist struct
type NewsRespons struct {
	Message string           `json:"message,omitempty"`
	Data    []domain.Article `json:"data,omitempty"`
}

// respons json
type Respons struct {
	Data Data `json:"data"`
}

type Data struct {
	Contents []Contents `json:"contents"`
}

type Contents struct {
	ID    string `json:"id"`
	Title Title  `json:"title"`
}

type Title struct {
	Short string `json:"short"`
}

func GetNewsList(q string) []string {
	payload := strings.NewReader(q)
	var list []string
	req, err := http.NewRequest("POST", Url, payload)
	if err != nil {
		fmt.Println("Wrapping request error", err)
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Requerst error, stoping program")
		panic(err)
	}

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	data := Respons{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println("Unmarshaling error GetNewsList", err)
	}

	for i := 0; i < len(data.Data.Contents); i++ {
		list = append(list, data.Data.Contents[i].ID)
		// fmt.Println("topic id = ", data.Data.Contents[i].ID)
		// fmt.Println("topic name = ", data.Data.Contents[i].Title.Short)
	}

	return list
}
