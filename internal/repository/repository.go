package repository

import (
	"github.com/FIFSAK/TMS/internal/domain/author"
	"github.com/FIFSAK/TMS/pkg/store"
)

type Configuration func(r *Repositories) error

type Repositories struct {
	postgres *store.SQL

	Author author.Repository
}

func New(configs ...Configuration) (s *Repositories, err error) {
	s = &Repositories{}

	for _, cfg := range configs {
		if err = cfg(s); err != nil {
			return
		}
	}

	return
}

func (r *Repositories) Close() {
	if r.postgres != nil && r.postgres.Connection != nil {
		r.postgres.Connection.Close()
	}
}

func WithSqlite(dataSourceName string) Configuration {
	return func(s *Repositories) (err error) {
		s.postgres, err = store.NewSQL(dataSourceName)
		if err != nil {
			return
		}

		if err = store.RunMigrations(dataSourceName); err != nil {
			return
		}
		//TODO finish
		return
	}
}
