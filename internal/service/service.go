package service

import (
	"github.com/FIFSAK/TMS/internal/config"
	"github.com/FIFSAK/TMS/internal/repository"
	"github.com/FIFSAK/TMS/internal/service/author"
)

type Dependencies struct {
	Repositories *repository.Repositories
	Configs      *config.Configs
}

type Configuration func(s *Services) error

type Services struct {
	dependencies Dependencies
	Author       author.Author
}

func New(dependencies Dependencies, configs ...Configuration) (s *Services, err error) {
	s = &Services{
		dependencies: dependencies,
	}

	for _, cfg := range configs {
		if err = cfg(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func WithLibraryService() Configuration {
	return func(s *Services) (err error) {

		s.Author = author.NewAuthorService(
			s.dependencies.Repositories.Author,
		)

		return nil
	}
}
