package usecases

import (
	"News/domain"
	"context"
)

type UseCases struct {
	articleRepo articleRepository
	searchRepo  searchRepository
}

func NewUseCases(ar articleRepository, sr searchRepository) *UseCases {
	return &UseCases{
		articleRepo: ar,
		searchRepo:  sr,
	}
}

// type articleUsecase interface {
// 	GetByID(ctx context.Context, id string) (domain.Article, error)
// }

type articleRepository interface {
	GetByID(ctx context.Context, id string) (domain.Article, error)
	GetNewsFromDB(ctx context.Context, take int, skip int) ([]domain.Article, error)
	UpdateOne(domain.Article) error
}

type searchRepository interface {
	Search(query string) ([]domain.Article, error)
	UpdateOne(domain.Article) error
}

func (uc *UseCases) GetByID(ctx context.Context, id string) (domain.Article, error) {
	return uc.articleRepo.GetByID(ctx, id)
}

func (uc *UseCases) GetArticlesList(ctx context.Context, take, skip int) ([]domain.Article, error) {
	return uc.articleRepo.GetNewsFromDB(ctx, take, skip)
}

func (uc *UseCases) EditArticle(art domain.Article) error {
	if err := uc.searchRepo.UpdateOne(art); err != nil {
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
