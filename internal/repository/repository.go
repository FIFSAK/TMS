package repository

import (
	"github.com/FIFSAK/TMS/internal/domain/shipment"
	"github.com/FIFSAK/TMS/internal/repository/postgres"
	"github.com/FIFSAK/TMS/pkg/store"
)

type Configuration func(r *Repositories) error

type Repositories struct {
	postgres *store.SQL

	Shipment shipment.Repository
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

func WithPostgresStore(dataSourceName string) Configuration {
	return func(s *Repositories) (err error) {
		s.postgres, err = store.NewSQL(dataSourceName)
		if err != nil {
			return
		}

		if err = store.RunMigrations(dataSourceName); err != nil {
			return
		}

		s.Shipment = postgres.NewShipmentRepository(s.postgres.Connection)

		return
	}
}
