package mysql

import (
	"github.com/d3z41k/url-shortener/shortener"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type mysqlRepository struct {
	client *gorm.DB
}

func newMysqlClient(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func NewMysqlRepository(dsn string) (shortener.RedirectRepository, error) {
	repo := &mysqlRepository{}
	client, err := newMysqlClient(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "repository.NewMysqlRepository")
	}
	repo.client = client
	return repo, nil
}

func (r *mysqlRepository) Find(code string) (*shortener.Redirect, error) {
	redirect := &shortener.Redirect{}

	if err := r.client.Where("code = ?", code).First(&redirect).Error; err != nil {
		return nil, errors.Wrap(err, "repository.Redirect.Find")
	}
	return redirect, nil
}

func (r *mysqlRepository) Store(redirect *shortener.Redirect) error {
	result := r.client.Create(&redirect)

	if result.Error != nil {
		return errors.Wrap(result.Error, "repository.Redirect.Store")
	}
	return nil
}
