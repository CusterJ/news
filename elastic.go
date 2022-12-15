package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var ES_ARTS = os.Getenv("ES_ARTS")

type Elastic struct {
	Url   string
	Index string
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
			Index   string   `json:"_index"`
			ID      string   `json:"_id"`
			Score   float64  `json:"_score"`
			Ignored []string `json:"_ignored"`
			Source  Article  `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (e *Elastic) ElasticReq() (*Elastic, error) {
	return &Elastic{
		Url:   os.Getenv("ES_URL"),
		Index: os.Getenv("ES_INDEX"),
	}, nil
}

// Mapping
func (e *Elastic) EsPutIndex() error {
	// !TODO mapping
	url, _ := e.ElasticReq()
	mapUrl := url.Url + url.Index + "_mapping/"
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

// CRUD
func EsInsertBulk(arts []Article) error {
	var data string
	for _, a := range arts {
		indexId := fmt.Sprintf(`{ "index": { "_id": "%s" }}`, a.Data.Content.Id)
		request, _ := json.Marshal(a)
		data += indexId + "\n" + string(request) + "\n"
	}
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

	// fmt.Println(res)
	// body, _ := io.ReadAll(res.Body)
	// fmt.Println(string(body))
	return nil
}

func EsInsertOne(art Article) error {
	return nil
}

func EsUpdateOne(art Article) error {
	return nil
}

func EsDeleteOne(id string) error {
	return nil
}

// SEARCH

func EsSearchArticle(s string) ([]Article, error) {
	fmt.Println("func EsSearchArticle start")
	data := fmt.Sprintf(`{
		"query": {
		  "match": {
			"data.content.title.short": "%s"
		  }
		}
	  }`, s)
	esres := &EsSearchResponse{}
	var arts []Article
	req, _ := http.NewRequest("GET", ES_ARTS+"_search", strings.NewReader(data))
	req.Header.Set("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), esres)
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
