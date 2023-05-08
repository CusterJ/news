package usecases

import (
	"News/domain"
	"context"
)

func (uc *UseCases) GetByID(ctx context.Context, id string) (domain.Article, error) {
	return uc.articleRepo.GetByID(ctx, id)
}

func (uc *UseCases) GetArticlesList(ctx context.Context, take, skip int) ([]domain.Article, error) {
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

func (uc *UseCases) Search(ctx context.Context, query string) ([]domain.Article, error) {
	return uc.searchRepo.Search(query)
}

func (uc *UseCases) Count(ctx context.Context) (int64, error) {
	docs, err := uc.articleRepo.Count(ctx)
	if err != nil {
		return 0, err
	}

	return docs, nil
}
