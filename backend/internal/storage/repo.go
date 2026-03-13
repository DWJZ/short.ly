package storage

import (
	"context"
	"crypto/rand"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repo interface {
	CreateShortURL(ctx context.Context, originalURL string) (code string, err error)
	ResolveOriginalURL(ctx context.Context, code string) (originalURL string, found bool, err error)
}

type GormRepo struct {
	db *gorm.DB
}

func OpenAndMigrate(databaseURL string) (*GormRepo, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.AutoMigrate(&OriginalURL{}, &ShortURL{}); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &GormRepo{db: db}, nil
}

func (r *GormRepo) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	var outCode string

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orig := OriginalURL{URL: originalURL}
		if err := tx.Create(&orig).Error; err != nil {
			return err
		}

		for i := 0; i < 5; i++ {
			code, err := newCode(8)
			if err != nil {
				return err
			}

			row := ShortURL{Code: code, OriginalURLID: orig.ID}
			if err := tx.Create(&row).Error; err != nil {
				// likely unique violation on code; retry
				continue
			}

			outCode = code
			return nil
		}

		return fmt.Errorf("could not allocate unique code")
	})

	if err != nil {
		return "", err
	}
	return outCode, nil
}

func (r *GormRepo) ResolveOriginalURL(ctx context.Context, code string) (string, bool, error) {
	var short ShortURL
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&short).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}

	var orig OriginalURL
	if err := r.db.WithContext(ctx).First(&orig, short.OriginalURLID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}

	return orig.URL, true, nil
}

const codeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func newCode(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("invalid code length")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}
	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(b), nil
}

