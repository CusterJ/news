package newsparser

import (
	"News/db/es"
	"News/db/mdb"
	"News/domain"
	"context"
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	mgo *mdb.ArticleRepo
	es  *es.ElasticRepo
}

func NewWorker(mgo *mdb.ArticleRepo, es *es.ElasticRepo) *Worker {
	return &Worker{
		mgo: mgo,
		es:  es,
	}
}

func (w *Worker) StartParser(ch chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ch:
			fmt.Println("Worker off")
			return
		case <-ticker.C:

			var articleRepoDate, firstNewsListDate int64 = 0, 0

			articleRepoDate, err := w.takeArticleRepoDate()
			if err != nil {
				// set one week if db is empty
				articleRepoDate = time.Now().Add(-time.Hour * 24).Unix()
				fmt.Printf("Parser -> can't take date from DB, set date: %v, humandate %v\n", articleRepoDate, time.Unix(articleRepoDate, 0))
			}

			firstNewsListDate, err = takeNewsListDate()
			if err != nil {
				fmt.Println("Parser -> can't take last news list date from site. Sleep and try later")
				time.Sleep(10 * time.Second)
				continue
			}

			// fmt.Printf("Parser start -> dates: \n repo = %v == %v \n news = %v == %v \n",
			// 	articleRepoDate, time.Unix(articleRepoDate, 0), firstNewsListDate, time.Unix(firstNewsListDate, 0))

			// compareDates(articleRepoDate ,firstNewsListDate)
			if articleRepoDate < firstNewsListDate {
				savedToDbs := 0
				for lastNewsListDate := firstNewsListDate + 10; articleRepoDate < lastNewsListDate; {
					news := getArticles(getNewsList(newsQuery(30, 0, lastNewsListDate)))
					fmt.Printf("Parser -> news, len = %v\n", len(news))
					// fmt.Printf("Parser -> news[0] %+v\n", news[0])

					err := w.saveArticlesToDBs(news)
					if err != nil {
						fmt.Println("Parser -> error saving news to dbs: ", err)
					}

					pd := news[len(news)-1].Dates.Posted
					if err != nil {
						fmt.Println("func Parser -> convert date to string Atoi for lastNewsListDate error", err)
						time.Sleep(1 * time.Minute)
						continue
					}
					lastNewsListDate = int64(pd)
					savedToDbs += len(news)

					fmt.Printf("Parser -> dates: repo = %v, news = %v, last post = %v, saved = %v \n",
						articleRepoDate, firstNewsListDate, lastNewsListDate, savedToDbs)
					fmt.Printf("Parser -> dates: repo = %v, news = %v, last post = %v, saved = %v \n",
						time.Unix(articleRepoDate, 0), time.Unix(firstNewsListDate, 0), time.Unix(lastNewsListDate, 0), savedToDbs)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

func takeNewsListDate() (int64, error) {
	date := time.Now().Unix()

	nd := getNewsListDate(newsQuery(1, 0, date))
	if nd == 0 {
		return 0, fmt.Errorf("func takeNewsListDate -> from func GetNewsListDate returned zero")
	}
	return nd, nil
}

func (w *Worker) saveArticlesToDBs(a []domain.Article) error {
	err := w.mgo.BulkWrite(a)
	if err != nil {
		fmt.Println("func SaveArticlesToDBs mongo save err: ", err)
		return err
	}
	err = w.es.EsInsertBulk(a)
	if err != nil {
		fmt.Println("func SaveArticlesToDBs es save rr: ", err)
		return err
	}
	return nil
}

func (w *Worker) takeArticleRepoDate() (int64, error) {
	// get article from MONGO
	var mongoDate int64

	mgoArticle, err := w.mgo.GetNewsFromDB(context.TODO(), 5, 0)
	if err != nil {
		return 0, err
	}

	if len(mgoArticle) > 0 {
		mongoDate = mgoArticle[0].Dates.Posted
		return mongoDate, nil
	}

	fmt.Printf("func newsparser.takeArticleRepoDate -> \n mgoDate: %v\n", mongoDate)
	return mongoDate, fmt.Errorf("can't take mongo date from articles")
}
