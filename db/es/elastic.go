package es

import (
	"News/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Articles struct {
	Articles *domain.Article
	Total    int
}

type ElasticRepo struct {
	url   string
	index string
}

type EsIndexArticle struct {
	IndexId struct {
		ID string `json:"index,omitempty"`
	} `json:"_id,omitempty"`
}

type EsSearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index   string         `json:"_index"`
			ID      string         `json:"_id"`
			Score   float64        `json:"_score"`
			Ignored []string       `json:"_ignored"`
			Source  domain.Article `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// ES take articles from page
func (a *Articles) GetPaginateResults(page int) (arts []domain.Article, err error) {
	size, err := strconv.Atoi(os.Getenv("TAKE"))
	if err != nil {
		fmt.Println("GetPaginateResults")
		return nil, err
	}
	from := 0
	if page > 0 {
		from = (size * page) - 1
	}
	payload := fmt.Sprintf(`{
		"from": %v,
		"size": %v
	  }`, from, size)

	ES_ARTS := os.Getenv("ES_ARTS")
	url := ES_ARTS + "_search"
	req, err := http.NewRequest("GET", url, strings.NewReader(payload))
	if err != nil {
		fmt.Println("func GetPaginateResults NewRequest error: ", err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("func GetPaginateResults Do(req) error: ", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	// fmt.Printf("func GetPaginateResults response from ES body: %s \n", body)
	esres := &EsSearchResponse{}
	err = json.Unmarshal(body, esres)
	if err != nil {
		fmt.Println("Unmarshaling GetPaginateResults error ", err)
		return
	}
	for _, v := range esres.Hits.Hits {
		// fmt.Printf("\n ---\n esres.Hits.Hits: %#v \n ---\n", v.Source.Data)
		arts = append(arts, v.Source)
	}
	return arts, nil
}

// ES count articles
func (a *Articles) Count() {
	fmt.Println("func EsCountArticles -> start")
	var count float64

	ES_ARTS := os.Getenv("ES_ARTS")
	url := ES_ARTS + "_count"
	es := make(map[string]interface{})
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("func EsCountArticles Get count error")
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("func Count ES *Articles -> read respons error")
		return
	}
	err = json.Unmarshal(body, &es)
	if err != nil {
		fmt.Println("func EsCountArticles Unmarshal response from ES error")
		return
	}
	count, ok := es["count"].(float64)
	if !ok {
		fmt.Println("func EsCountArticles type assertion response from ES error")
		return
	}
	fmt.Printf("func EsCountArticles -> COUNT = %v \n", count)
	a.Total = int(count)
}

// CRUD
func EsInsertBulk(arts []domain.Article) error {
	var data string
	for _, a := range arts {
		indexId := fmt.Sprintf(`{ "index": { "_id": "%s" }}`, a.Data.Content.Id)
		request, _ := json.Marshal(a)
		data += indexId + "\n" + string(request) + "\n"
	}

	ES_ARTS := os.Getenv("ES_ARTS")
	url := ES_ARTS + "_bulk"
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		fmt.Println("func EsInsertBulk: create new req error")
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("func EsInsertBulk: Requerst error")
		return err
	}
	fmt.Println("func EsInsertBulk: ES respons", res.StatusCode)
	defer res.Body.Close()

	// body, _ := io.ReadAll(res.Body)
	// fmt.Println(string(body))
	return nil
}

func EsInsertOne(art domain.Article) error {
	return nil
}

func EsUpdateOne(art domain.Article) error {
	ES_ARTS := os.Getenv("ES_ARTS")
	url := ES_ARTS + "_doc/" + art.Data.Content.Id

	body, err := json.Marshal(art)
	if err != nil {
		fmt.Println("EsUpdateOne json.Marshal error", err)
		return err
	}

	b := strings.NewReader(string(body))
	req, err := http.NewRequest("PUT", url, b)
	if err != nil {
		fmt.Println("EsUpdateOne NewReader error", err)
		return err
	}
	fmt.Printf("\n func EsUpdateOne --> URL %v, \n BODY %v \n", url, b)
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		// Aditional add check response that  "result": "updated"
		fmt.Println("func EsUpdateOne DO request error")
		return err
	}
	resb, _ := io.ReadAll(res.Body)
	fmt.Println(string(resb))
	return nil
}

func EsDeleteOne(id string) error {
	return nil
}

// SEARCH

func EsSearchArticle(s string) (arts []domain.Article, err error) {
	fmt.Println("func EsSearchArticle start")
	ES_ARTS := os.Getenv("ES_ARTS")

	esres := &EsSearchResponse{}

	data := fmt.Sprintf(`{
		"query": {
		  "multi_match": {
			"query": "%s",
			"fields": [
			  "data.content.title.short",
			  "data.content.description.long"
			]
		  }
		}
	  }`, s)

	req, _ := http.NewRequest("GET", ES_ARTS+"_search", strings.NewReader(data))
	req.Header.Set("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	err = json.Unmarshal(body, esres)
	if err != nil {
		fmt.Println("Unmarshaling EsSearchArticle error ", err)
		return nil, err
	}

	for _, v := range esres.Hits.Hits {
		// fmt.Printf("\n ---\n esres.Hits.Hits: %#v \n ---\n", v.Source.Data)
		arts = append(arts, v.Source)
	}
	// fmt.Println("Print arts: ", arts)
	return arts, nil
}

func (e *ElasticRepo) ElasticReq() (*ElasticRepo, error) {
	return &ElasticRepo{
		url:   os.Getenv("ES_URL"),
		index: os.Getenv("ES_INDEX"),
	}, nil
}

// ES create index
func EsCreateIndex(index string) error {
	return nil
}

// Mapping
func (e *ElasticRepo) EsPutIndex() error {
	// !TODO mapping
	url, _ := e.ElasticReq()
	mapUrl := url.url + url.index + "_mapping/"
	// {
	// 	"settings": {
	// 		"number_of_shards": 1,
	// 		"number_of_replicas": 1
	// 	},
	//    "mappings": {
	// 	   "properties": {
	// 		 "name": {
	// 			   "type": "text"
	// 		 },
	// 		 "age": {
	// 			   "type": "integer"
	// 		 },
	// 		 "average_score": {
	// 			   "type": "float"
	// 		 }
	// 	 }
	//    }
	// }

	mp := `{
		"properties": {
		  "content": {
			"title": {
			  "short": {
				"type": "text"
			  }
			},
			"description": {
			  "long": {
				"type": "text"
			  }
			}
		  }
		}
	  }`

	req, _ := http.NewRequest("PUT", mapUrl, strings.NewReader(mp))
	http.DefaultClient.Do(req)
	return nil
}
