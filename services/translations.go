package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pgrzankowski/dictionary-app/graph/model"
	gormModels "github.com/pgrzankowski/dictionary-app/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateTranslation(db *gorm.DB, ctx context.Context, input model.NewTranslationInput) (*model.Translation, error) {
	transaction := db.Begin()
	if transaction.Error != nil {
		return nil, transaction.Error
	}

	var polishWord gormModels.PolishWord

	if err := transaction.Clauses(clause.OnConflict{DoNothing: true}).Create(&gormModels.PolishWord{Word: input.PolishWord}).Error; err != nil {
		transaction.Rollback()
		return nil, fmt.Errorf("failed to create polish word: %w", err)
	}

	if err := transaction.Where("word = ?", input.PolishWord).First(&polishWord).Error; err != nil {
		transaction.Rollback()
		return nil, fmt.Errorf("failed to fetch polish word: %w", err)
	}

	var existingTranslation gormModels.Translation
	if err := transaction.
		Where("polish_word_id = ? AND english_word = ?", polishWord.ID, input.EnglishWord).
		First(&existingTranslation).Error; err == nil {
		transaction.Rollback()
		return nil, fmt.Errorf("translation for polish word '%s' with english word '%s' already exists", input.PolishWord, input.EnglishWord)
	} else if err != gorm.ErrRecordNotFound {
		transaction.Rollback()
		return nil, fmt.Errorf("error checking for existing translation: %w", err)
	}

	translation := gormModels.Translation{
		EnglishWord:  input.EnglishWord,
		PolishWordID: polishWord.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := transaction.Create(&translation).Error; err != nil {
		transaction.Rollback()
		return nil, fmt.Errorf("failed to create translation: %w", err)
	}

	for _, exInput := range input.Examples {
		example := gormModels.Example{
			Sentence:      exInput.Sentence,
			TranslationID: translation.ID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := transaction.Create(&example).Error; err != nil {
			transaction.Rollback()
			return nil, fmt.Errorf("failed to create example: %w", err)
		}
		translation.Examples = append(translation.Examples, example)
	}

	if err := transaction.Commit().Error; err != nil {
		return nil, err
	}

	transaction.Model(&translation).Association("PolishWord").Find(&translation.PolishWord)

	return &model.Translation{
		ID:          strconv.Itoa(int(translation.ID)),
		EnglishWord: translation.EnglishWord,
		CreatedAt:   translation.CreatedAt.String(),
		UpdatedAt:   translation.UpdatedAt.String(),
		PolishWord: &model.PolishWord{
			ID:        strconv.Itoa(int(polishWord.ID)),
			Word:      polishWord.Word,
			CreatedAt: polishWord.CreatedAt.String(),
			UpdatedAt: polishWord.UpdatedAt.String(),
		},
		Examples: convertExamples(translation.Examples),
	}, nil
}

func RemoveTranslation(db *gorm.DB, ctx context.Context, id string) (bool, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return false, fmt.Errorf("invalid id format: %w", err)
	}

	transaction := db.Begin()
	if transaction.Error != nil {
		return false, transaction.Error
	}

	var translation gormModels.Translation
	if err := transaction.
		Preload("PolishWord").
		First(&translation, intID).
		Error; err != nil {
		transaction.Rollback()
		return false, fmt.Errorf("failed to fetch translation: %w", err)
	}

	polishWordID := translation.PolishWordID

	if err := transaction.
		Delete(&translation).Error; err != nil {
		transaction.Rollback()
		return false, fmt.Errorf("failed to delete translation: %w", err)
	}

	var translationCount int64
	if err := transaction.Model(&gormModels.Translation{}).
		Where("polish_word_id = ?", polishWordID).
		Count(&translationCount).Error; err != nil {
		transaction.Rollback()
		return false, fmt.Errorf("failed to count translations: %w", err)
	}

	if translationCount == 0 {
		if err := transaction.
			Delete(&gormModels.PolishWord{}, polishWordID).
			Error; err != nil {
			transaction.Rollback()
			return false, fmt.Errorf("failed to delete polish word: %w", err)
		}
	}

	if err := transaction.Commit().Error; err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return true, nil
}

func UpdateTranslation(db *gorm.DB, ctx context.Context, input model.UpdateTranslationInput) (*model.Translation, error) {
	intID, err := strconv.Atoi(input.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %v", err)
	}

	transaction := db.Begin()
	if transaction.Error != nil {
		return nil, transaction.Error
	}

	var translation gormModels.Translation
	if err := transaction.
		Preload("PolishWord").
		Preload("Examples").
		First(&translation, intID).
		Error; err != nil {
		transaction.Rollback()
		return nil, err
	}

	translation.EnglishWord = *input.EnglishWord
	translation.UpdatedAt = time.Now()

	if err := transaction.Save(&translation).Error; err != nil {
		transaction.Rollback()
		return nil, err
	}

	if err := transaction.Commit().Error; err != nil {
		return nil, err
	}

	updatedTranslation := &model.Translation{
		ID:          strconv.Itoa(int(translation.ID)),
		EnglishWord: translation.EnglishWord,
		CreatedAt:   translation.EnglishWord,
		UpdatedAt:   translation.UpdatedAt.String(),
		PolishWord: &model.PolishWord{
			ID:        strconv.Itoa(int(translation.PolishWord.ID)),
			Word:      translation.PolishWord.Word,
			CreatedAt: translation.PolishWord.CreatedAt.String(),
			UpdatedAt: translation.PolishWord.UpdatedAt.String(),
		},
		Examples: convertExamples(translation.Examples),
	}

	return updatedTranslation, nil
}

func Translations(db *gorm.DB, ctx context.Context) ([]*model.Translation, error) {
	var translations []gormModels.Translation
	if err := db.
		Preload("PolishWord").
		Preload("Examples").
		Find(&translations).Error; err != nil {
		return nil, err
	}

	var result []*model.Translation
	for _, translation := range translations {
		result = append(result, &model.Translation{
			ID:          strconv.Itoa(int(translation.ID)),
			EnglishWord: translation.EnglishWord,
			CreatedAt:   translation.CreatedAt.String(),
			UpdatedAt:   translation.UpdatedAt.String(),
			PolishWord: &model.PolishWord{
				ID:        strconv.Itoa(int(translation.PolishWord.ID)),
				Word:      translation.PolishWord.Word,
				CreatedAt: translation.PolishWord.CreatedAt.String(),
				UpdatedAt: translation.PolishWord.UpdatedAt.String(),
			},
			Examples: convertExamples(translation.Examples),
		})
	}

	return result, nil
}

func Translation(db *gorm.DB, ctx context.Context, id string) (*model.Translation, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %v", err)
	}

	var translation gormModels.Translation
	if err := db.
		Preload("PolishWord").
		Preload("Examples").
		First(&translation, intID).Error; err != nil {
		return nil, err
	}

	result := &model.Translation{
		ID:          strconv.Itoa(int(translation.ID)),
		EnglishWord: translation.EnglishWord,
		CreatedAt:   translation.CreatedAt.String(),
		UpdatedAt:   translation.UpdatedAt.String(),
		PolishWord: &model.PolishWord{
			ID:        strconv.Itoa(int(translation.PolishWord.ID)),
			Word:      translation.PolishWord.Word,
			CreatedAt: translation.PolishWord.CreatedAt.String(),
			UpdatedAt: translation.PolishWord.UpdatedAt.String(),
		},
		Examples: convertExamples(translation.Examples),
	}

	return result, nil
}

func convertExamples(examples []gormModels.Example) []*model.Example {
	var result []*model.Example
	for _, ex := range examples {
		result = append(result, &model.Example{
			ID:        strconv.Itoa(int(ex.ID)),
			Sentence:  ex.Sentence,
			CreatedAt: ex.CreatedAt.String(),
			UpdatedAt: ex.UpdatedAt.String(),
		})
	}
	return result
}
