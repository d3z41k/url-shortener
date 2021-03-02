package sqlite

import (
	"github.com/d3z41k/url-shortener/shortener"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

type sqliteRepository struct {
	client *gorm.DB
}

func newSqliteClient(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewSqliteRepository(dsn string) (shortener.RedirectRepository, error) {
	repo := &sqliteRepository{}
	client, err := newSqliteClient(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "repository.NewSqliteRepository")
	}

	log.Println(dsn)
	repo.client = client
	return repo, nil
}

func (r *sqliteRepository) Find(code string) (*shortener.Redirect, error) {
	redirect := &shortener.Redirect{}

	if err := r.client.Where("code = ?", code).First(&redirect).Error; err != nil {
		return nil, errors.Wrap(err, "repository.Redirect.Find")
	}
	return redirect, nil
}

func (r *sqliteRepository) Store(redirect *shortener.Redirect) error {
	result := r.client.Create(&redirect)

	if result.Error != nil {
		return errors.Wrap(result.Error, "repository.Redirect.Store")
	}
	return nil
}
