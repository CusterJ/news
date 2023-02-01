package newsparser

import (
	"encoding/json"
	"fmt"
)

// Query for get list of articles IDs
type PointQuery struct {
	Query     string  `json:"query,omitempty"`
	Variables VarNews `json:"variables,omitempty"`
}

type VarNews struct {
	ProjectId string `json:"projectId"`
	Lang      string `json:"lang"`
	Take      int    `json:"take"`
	Skip      int    `json:"skip"`
	DateTo    int64  `json:"dateTo"`
}

// Query for get one article by ID
type ArtQuery struct {
	Query     string     `json:"query,omitempty"`
	Variables VarArticle `json:"variables,omitempty"`
}

type VarArticle struct {
	Id string `json:"id,omitempty"`
}

func newsQuery(take int, skip int, dateTo int64) string {
	q := &PointQuery{
		Query: "query contents($projectId: String!, $lang: String = \"ru\", $take: Int = 30, $skip: Int, $dateTo: Int) {\n  contents(\n    project_id: $projectId\n    lang: $lang\n    take: $take\n    skip: $skip\n    posted_date_to: $dateTo\n  ) {\n    id\n    title {\n      short\n    }\n dates {\n      posted\n  }\n }\n}\n",
		Variables: VarNews{
			Lang:      "ru",
			Take:      take,
			ProjectId: "5107de83-f208-4ca4-87ed-9b69d58d16e1",
			Skip:      skip,
			DateTo:    dateTo,
		},
	}
	query, err := json.Marshal(q)
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(query)
}

func articleQuery(id string) string {
	q := &ArtQuery{
		Query: "query content($id: String!) {\n  content(\n    id: $id\n  ) {\n\turl\n \n id \n   title {\n      short\n    }  \n    description {\n        long\n    } \n   dates {\n      posted\n    } \n}\n}",
		Variables: VarArticle{
			Id: id,
		},
	}
	query, err := json.Marshal(q)
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(query)
}
