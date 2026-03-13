package storage

import "time"

type OriginalURL struct {
	ID        uint      `gorm:"primaryKey"`
	URL       string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (OriginalURL) TableName() string { return "original_urls" }

type ShortURL struct {
	ID            uint      `gorm:"primaryKey"`
	Code          string    `gorm:"not null;uniqueIndex"`
	OriginalURLID uint      `gorm:"not null;index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (ShortURL) TableName() string { return "short_urls" }

