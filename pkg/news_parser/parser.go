package newsparser

import (
	"News/db/es"
	"News/db/mdb"
	"News/domain"
	"context"
	"fmt"
	"log"
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
	ticker := time.NewTicker(5 * time.Second)
	take, skip := 30, 0

	defer wg.Done()

	for {
		select {
		case <-ch:
			fmt.Println("Worker off")

			return
		case <-ticker.C:
			var articleRepoDate, firstNewsListDate int64

			articleRepoDate, err := w.takeArticleRepoDate()
			if err != nil {
				// get news from if db is empty
				articleRepoDate = time.Now().Add(-time.Hour * 24 * 7).Unix()
				log.Printf("Parser -> can't take date from DB, set date: %v, humandate %v\n",
					articleRepoDate, time.Unix(articleRepoDate, 0))
			}

			firstNewsListDate, err = takeNewsListDate()
			if err != nil {
				log.Println("Parser -> can't take last news list date from site. Sleep and try later")
				log.Println("Parser -> error: ", err)
				time.Sleep(10 * time.Second)

				continue // break
			}

			if articleRepoDate < firstNewsListDate {
				savedToDbs := 0

				for lastNewsListDate := firstNewsListDate + 10; articleRepoDate < lastNewsListDate; {
					articlesList, err := getNewsList(newsQuery(take, skip, lastNewsListDate))
					if err != nil {
						log.Printf("Parser -> get list of articles from site error: %s. Sleep for minute and continue\n", err)

						time.Sleep(time.Minute)

						continue // break?
					}

					news := getArticles(articlesList)

					err = w.saveArticlesToDBs(news)
					if err != nil {
						log.Println("Parser -> error saving news to dbs: ", err)
						// break or continue?
					}

					pd := news[len(news)-1].Dates.Posted

					if err != nil {
						log.Println("func Parser -> take last news list date error", err)
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

	nd, err := getNewsListDate(newsQuery(1, 0, date))
	if err != nil || nd == 0 {
		return 0, fmt.Errorf("func takeNewsListDate -> from func GetNewsListDate returned ZERO or error: %s", err)
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
		log.Println("func SaveArticlesToDBs es save rr: ", err)

		return err
	}

	return nil
}

// Get article date from MONGO.
func (w *Worker) takeArticleRepoDate() (int64, error) {
	var mongoDate int64

	mgoArticle, err := w.mgo.GetNewsFromDB(context.TODO(), 1, 0)
	if err != nil {
		return 0, err
	}

	if len(mgoArticle) > 0 {
		mongoDate = mgoArticle[0].Dates.Posted

		return mongoDate, nil
	}

	log.Printf("func newsparser.takeArticleRepoDate -> \n mgoDate: %v\n", mongoDate)

	return mongoDate, fmt.Errorf("can't take mongo date from articles")
}
