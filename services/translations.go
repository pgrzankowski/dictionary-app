package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pgrzankowski/dictionary-app/db"
	"github.com/pgrzankowski/dictionary-app/graph/model"
	gormModels "github.com/pgrzankowski/dictionary-app/models"
)

func CreateTranslation(ctx context.Context, input model.NewTranslationInput) (*model.Translation, error) {
	transaction := db.GormDB.Begin()
	if transaction.Error != nil {
		return nil, transaction.Error
	}

	var polishWord gormModels.PolishWord
	if err := transaction.Where("word = ?", input.PolishWord).
		FirstOrCreate(&polishWord, gormModels.PolishWord{Word: input.PolishWord}).
		Error; err != nil {
		transaction.Rollback()
		return nil, fmt.Errorf("failed to get or create polish word: %w", err)
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

func RemoveTranslation(ctx context.Context, id string) (bool, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return false, fmt.Errorf("invalid id format: %w", err)
	}

	transaction := db.GormDB.Begin()
	if transaction.Error != nil {
		return false, transaction.Error
	}

	var translation gormModels.Translation
	if err := transaction.
		Preload("PolishWord").
		First(&translation, intID).
		Error; err != nil {
		transaction.Rollback()
		return false, err
	}

	polishWordID := translation.PolishWordID

	if err := transaction.Delete(&translation).Error; err != nil {
		transaction.Rollback()
		return false, err
	}

	var translationCount int64
	if err := transaction.Model(&gormModels.Translation{}).
		Where("polish_word_id = ?", polishWordID).
		Count(&translationCount).Error; err != nil {
		transaction.Rollback()
		return false, err
	}

	if translationCount == 0 {
		if err := transaction.Delete(&gormModels.PolishWord{}, polishWordID).
			Error; err != nil {
			transaction.Rollback()
			return false, nil
		}
	}

	if err := transaction.Commit().Error; err != nil {
		return false, err
	}

	return true, nil
}

func UpdateTranslation(ctx context.Context, input model.UpdateTranslationInput) (*model.Translation, error) {
	intID, err := strconv.Atoi(input.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %v", err)
	}

	transaction := db.GormDB.Begin()
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

func Translations(ctx context.Context) ([]*model.Translation, error) {
	var translations []gormModels.Translation
	if err := db.GormDB.
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

func Translation(ctx context.Context, id string) (*model.Translation, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %v", err)
	}

	var translation gormModels.Translation
	if err := db.GormDB.
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
