package models

import (
	"time"
)

type PolishWord struct {
	ID           uint   `gorm:"primaryKey"`
	Word         string `gorm:"not null;uniqueIndex:idx_polish_word"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Translations []Translation `gorm:"foreignKey:PolishWordID"`
}

func (PolishWord) TableName() string {
	return "polish_words"
}

type Translation struct {
	ID           uint   `gorm:"primaryKey"`
	PolishWordID uint   `gorm:"not null"`
	EnglishWord  string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PolishWord   PolishWord
	Examples     []Example `gorm:"foreignKey:TranslationID;constraint:OnDelete:CASCADE;"`
}

func (Translation) TableName() string {
	return "translations"
}

type Example struct {
	ID            uint   `gorm:"primaryKey"`
	TranslationID uint   `gorm:"not null"`
	Sentence      string `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Translation   Translation
}

func (Example) TableName() string {
	return "examples"
}
