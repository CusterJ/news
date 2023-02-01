package newsparser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const Url string = "https://point.md/graphql"

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
	Dates Dates  `json:"dates"`
}

type Title struct {
	Short string `json:"short"`
}

type Dates struct {
	Posted string `json:"posted"`
}

func getNewsList(q string) []string {
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
	}

	return list
}

func getNewsListDate(q string) int64 {
	var date int
	payload := strings.NewReader(q)
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
		fmt.Println("Unmarshaling error GetNewsListDate", err)
	}

	if len(data.Data.Contents) > 0 {
		date, err = strconv.Atoi(data.Data.Contents[0].Dates.Posted)
	}
	if err != nil {
		fmt.Println("func GetNewsListDate -> Atoi date error")
	}

	return int64(date)
}
