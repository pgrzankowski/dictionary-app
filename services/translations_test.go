package services_test

import (
	"context"
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/pgrzankowski/dictionary-app/db"
	"github.com/pgrzankowski/dictionary-app/graph/model"
	"github.com/pgrzankowski/dictionary-app/services"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// Clear test db
func clearTestDB(t *testing.T) {
	err := db.GormTestDB.Exec("TRUNCATE TABLE examples, translations, polish_words RESTART IDENTITY CASCADE").Error
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

// Test mutations
func TestCreateTranslation(t *testing.T) {

	db.ConnectTestGORM()
	clearTestDB(t)

	input := model.NewTranslationInput{
		PolishWord:  "pisać",
		EnglishWord: "write",
		Examples: []*model.NewExampleInput{
			{Sentence: "On lubi pisać listy."},
		},
	}

	ctx := context.Background()
	translation, err := services.CreateTranslation(db.GormTestDB, ctx, input)
	assert.NoError(t, err, "CreateTranslation should not return an error")
	assert.NotNil(t, translation, "translation should not be nil")
	assert.Equal(t, "write", translation.EnglishWord, "EnglishWord should match")
	assert.Equal(t, "pisać", translation.PolishWord.Word, "PolishWord should match")
	assert.Equal(t, 1, len(translation.Examples), "Examples list should have 1 element")
	assert.Equal(t, "On lubi pisać listy.", translation.Examples[0].Sentence, "Sentence should match")
}

func TestRemoveTranslation(t *testing.T) {

	db.ConnectTestGORM()
	clearTestDB(t)

	input := model.NewTranslationInput{
		PolishWord:  "pisać",
		EnglishWord: "write",
		Examples: []*model.NewExampleInput{
			{Sentence: "On lubi pisać listy."},
		},
	}

	ctx := context.Background()
	translation, _ := services.CreateTranslation(db.GormTestDB, ctx, input) // Already tested
	removed, err := services.RemoveTranslation(db.GormTestDB, ctx, translation.ID)
	assert.NoError(t, err, "RemoveTranslation should not return an error")
	assert.NotNil(t, removed, "removed should not be nil")
	assert.True(t, removed, "removed should be true")
}

func TestUpdateTranslation(t *testing.T) {

	db.ConnectTestGORM()
	clearTestDB(t)

	createInput := model.NewTranslationInput{
		PolishWord:  "pisać",
		EnglishWord: "write",
		Examples: []*model.NewExampleInput{
			{Sentence: "On lubi pisać listy."},
		},
	}

	ctx := context.Background()
	translation, _ := services.CreateTranslation(db.GormTestDB, ctx, createInput)

	updatedEnglishWord := "type"
	updateInput := model.UpdateTranslationInput{
		ID:          translation.ID,
		EnglishWord: &updatedEnglishWord,
	}

	updatedTranslation, err := services.UpdateTranslation(db.GormTestDB, ctx, updateInput)
	assert.NoError(t, err, "UpdateTranslation should not return an error")
	assert.NotNil(t, updatedTranslation, "updatedTranslation should not be nil")
	assert.Equal(t, "type", updatedTranslation.EnglishWord, "EnglishWord should match")
	assert.Equal(t, "pisać", updatedTranslation.PolishWord.Word, "PolishWord should match")
}

// Test queries
func TestTranslations(t *testing.T) {

	db.ConnectTestGORM()
	clearTestDB(t)

	inputTranslations := []model.NewTranslationInput{
		{
			PolishWord:  "pisać",
			EnglishWord: "write",
			Examples: []*model.NewExampleInput{
				{Sentence: "On lubi pisać listy."},
			},
		},
		{
			PolishWord:  "pić",
			EnglishWord: "drink",
			Examples: []*model.NewExampleInput{
				{Sentence: "On lubi pić wodę."},
				{Sentence: "Lubi też pić kawę."},
			},
		},
		{
			PolishWord:  "jeść",
			EnglishWord: "eat",
			Examples: []*model.NewExampleInput{
				{Sentence: "On lubi jeść pizze."},
				{Sentence: "Ona nie lubi jeść pizzy."},
				{Sentence: "Oni kochają jeść mięso."},
			},
		},
	}

	ctx := context.Background()

	var createdTranslations []*model.Translation

	for _, translation := range inputTranslations {
		createdTranslation, _ := services.CreateTranslation(db.GormTestDB, ctx, translation)
		createdTranslations = append(createdTranslations, createdTranslation)
	}

	translations, err := services.Translations(db.GormTestDB, ctx)

	assert.NoError(t, err, "Translations should not return an error")
	assert.NotNil(t, translations)
	assert.Equal(t, 3, len(translations), "Translations lenght should match")

	for ix, translation := range translations {
		assert.Equal(t, createdTranslations[ix].ID, translation.ID, "ID should match")
		assert.Equal(t, createdTranslations[ix].PolishWord.Word, translation.PolishWord.Word, "PolishWord should match")
		assert.Equal(t, createdTranslations[ix].EnglishWord, translation.EnglishWord, "EnglishWord should match")
		for idx, sentence := range translation.Examples {
			assert.Equal(t, createdTranslations[ix].Examples[idx].Sentence, sentence.Sentence, "Sentence should match")
		}
	}
}

func TestTranslation(t *testing.T) {

	db.ConnectTestGORM()
	clearTestDB(t)

	input := model.NewTranslationInput{
		PolishWord:  "pisać",
		EnglishWord: "write",
		Examples: []*model.NewExampleInput{
			{Sentence: "On lubi pisać listy."},
		},
	}

	ctx := context.Background()
	createdTranslation, _ := services.CreateTranslation(db.GormTestDB, ctx, input)

	translation, err := services.Translation(db.GormTestDB, ctx, createdTranslation.ID)
	assert.NoError(t, err, "Translation should not return an error")
	assert.NotNil(t, translation, "translation should not be nil")
	assert.Equal(t, createdTranslation.ID, translation.ID, "ID should match")
	assert.Equal(t, createdTranslation.EnglishWord, translation.EnglishWord, "EnglishWord should match")
	assert.Equal(t, createdTranslation.PolishWord.ID, translation.PolishWord.ID, "PolishWord.ID should match")
	assert.Equal(t, createdTranslation.PolishWord.Word, translation.PolishWord.Word, "PolishWord.Word should match")
	for ix, example := range translation.Examples {
		assert.Equal(t, createdTranslation.Examples[ix].Sentence, example.Sentence, "Sentences should match")
	}
}

// Test concurency
