package author

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/FIFSAK/TMS/internal/domain/author"
	"github.com/FIFSAK/TMS/pkg/log"
	"github.com/FIFSAK/TMS/pkg/store"
)

type AuthorService struct {
	authorRepository author.Repository
}

func NewAuthorService(r author.Repository) *AuthorService {
	return &AuthorService{
		authorRepository: r,
	}
}

type Author interface {
	ListAuthors(ctx context.Context) ([]author.Response, error)
	AddAuthor(ctx context.Context, req author.Request) (author.Response, error)
	GetAuthor(ctx context.Context, id string) (author.Response, error)
	UpdateAuthor(ctx context.Context, id string, req author.Request) error
	DeleteAuthor(ctx context.Context, id string) error
}

func (s *AuthorService) ListAuthors(ctx context.Context) ([]author.Response, error) {
	logger := log.FromContext(ctx).Named("list_authors")

	authors, err := s.authorRepository.List(ctx)
	if err != nil {
		logger.Error("failed to list authors", zap.Error(err))
		return nil, err
	}
	return author.ParseFromEntities(authors), nil
}

func (s *AuthorService) AddAuthor(ctx context.Context, req author.Request) (author.Response, error) {
	logger := log.FromContext(ctx).Named("add_author").With(zap.Any("author", req))

	newAuthor := author.New(req)

	id, err := s.authorRepository.Add(ctx, newAuthor)
	if err != nil {
		logger.Error("failed to add author", zap.Error(err))
		return author.Response{}, err
	}
	newAuthor.ID = id

	return author.ParseFromEntity(newAuthor), nil
}

func (s *AuthorService) GetAuthor(ctx context.Context, id string) (author.Response, error) {
	logger := log.FromContext(ctx).Named("get_author").With(zap.String("id", id))

	repoAuthor, err := s.authorRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("author not found", zap.Error(err))
			return author.Response{}, err
		}
		logger.Error("failed to get author", zap.Error(err))
		return author.Response{}, err
	}

	return author.ParseFromEntity(repoAuthor), nil
}

func (s *AuthorService) UpdateAuthor(ctx context.Context, id string, req author.Request) error {
	logger := log.FromContext(ctx).Named("update_author").With(zap.String("id", id), zap.Any("author", req))

	updatedAuthor := author.New(req)

	err := s.authorRepository.Update(ctx, id, updatedAuthor)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("author not found", zap.Error(err))
			return err
		}
		logger.Error("failed to update author", zap.Error(err))
		return err
	}

	return nil
}

func (s *AuthorService) DeleteAuthor(ctx context.Context, id string) error {
	logger := log.FromContext(ctx).Named("delete_author").With(zap.String("id", id))

	err := s.authorRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("author not found", zap.Error(err))
			return err
		}
		logger.Error("failed to delete author", zap.Error(err))
		return err
	}

	return nil
}
