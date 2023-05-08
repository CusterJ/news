package usecases

import (
	"News/domain"
	"context"
)

type UseCases struct {
	articleRepo articleRepository
	searchRepo  searchRepository
	userRepo    userRepository
}

func NewUseCases(ar articleRepository, ur userRepository, sr searchRepository) *UseCases {
	return &UseCases{
		articleRepo: ar,
		searchRepo:  sr,
		userRepo:    ur,
	}
}

// type repository interface {
// 	articleRepository
// 	searchRepository
// 	userRepository
// }

type articleRepository interface {
	GetByID(context.Context, string) (domain.Article, error)
	GetNewsFromDB(context.Context, int, int) ([]domain.Article, error)
	UpdateOne(domain.Article) error
	Count(context.Context) (int64, error)
}

type searchRepository interface {
	Search(string, int, int) ([]domain.Article, int, error)
	UpdateOne(domain.Article) error
}

type userRepository interface {
	UserSave(domain.User) error
	UserExistsInDB(string) (domain.User, bool)
	UserFind(string, string) error
}
