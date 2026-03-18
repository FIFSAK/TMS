package sqlite

import (
	"context"
	_ "errors"

	"github.com/mattn/go-sqlite3"

	"github.com/FIFSAK/TMS/internal/domain/author"
	_ "github.com/FIFSAK/TMS/pkg/store"
)

// TODO: implement requests
type AuthorRepository struct {
	db *sqlite3.SQLiteDriver
}

func NewAuthorRepository(db *sqlite3.SQLiteDriver) *AuthorRepository {
	return &AuthorRepository{
		db: db,
	}
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Entity, error) {
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {

}

func (r *AuthorRepository) Get(ctx context.Context, id string) (author.Entity, error) {

}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) error {

}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {

}
