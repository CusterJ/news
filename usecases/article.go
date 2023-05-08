package usecases

import (
	"News/domain"
	"context"
	"os"
	"strconv"
)

func (uc *UseCases) GetByID(ctx context.Context, id string) (domain.Article, error) {
	return uc.articleRepo.GetByID(ctx, id)
}

func (uc *UseCases) GetArticlesList(ctx context.Context, page int) ([]domain.Article, error) {
	skip := 0
	take, err := strconv.Atoi(os.Getenv("TAKE"))
	if err != nil {
		take = 15 // set default
	}

	if page >= 2 {
		skip = (page - 1) * take
	}

	return uc.articleRepo.GetNewsFromDB(ctx, take, skip)
}

func (uc *UseCases) EditArticle(art domain.Article) error {
	if err := uc.articleRepo.UpdateOne(art); err != nil {
		return err
	}
	if err := uc.searchRepo.UpdateOne(art); err != nil {
		return err
	}
	return nil
}

func (uc *UseCases) Search(ctx context.Context, query string, page int) (arts []domain.Article, hits int, err error) {
	skip := 0
	take, err := strconv.Atoi(os.Getenv("TAKE"))
	if err != nil {
		take = 15 // set default
	}

	if page >= 2 {
		skip = (page - 1) * take
	}

	return uc.searchRepo.Search(query, take, skip)
}

func (uc *UseCases) Count(ctx context.Context) (int64, error) {
	docs, err := uc.articleRepo.Count(ctx)
	if err != nil {
		return 0, err
	}

	return docs, nil
}
