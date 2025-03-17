package services_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

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
	translation, _ := services.CreateTranslation(db.GormTestDB, ctx, input)
	removed, err := services.RemoveTranslation(db.GormTestDB, ctx, translation.ID)
	assert.NoError(t, err, "RemoveTranslation should not return an error")
	assert.NotNil(t, removed, "removed should not be nil")
	assert.True(t, removed, "removed should be true")

	_, err = services.Translation(db.GormTestDB, ctx, translation.ID)
	assert.Error(t, err, "Quering deleted translation should return an error")
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
func TestConcurrentCreateTranslation(t *testing.T) {

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

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	start := false
	var wg sync.WaitGroup

	var results []*model.Translation
	var errors []error

	const iterations = 50
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			for !start {
				cond.Wait()
			}
			mu.Unlock()
			result, err := services.CreateTranslation(db.GormTestDB, ctx, input)
			mu.Lock()
			results = append(results, result)
			errors = append(errors, err)
			mu.Unlock()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	start = true
	cond.Broadcast()
	mu.Unlock()

	wg.Wait()

	var successCount, errorCount int
	var successIdx int
	for i := 0; i < iterations; i++ {
		if results[i] != nil {
			successCount++
			successIdx = i
		}
		if errors[i] != nil {
			assert.Contains(t, errors[i].Error(), "already exists", fmt.Sprintf("expected duplicate error, got: %v", errors[i]))
			errorCount++
		}
	}

	assert.Equal(t, 1, successCount, "Expected exactly one successful creation")
	assert.Equal(t, iterations-1, errorCount, "Expected the rest of the creations to fail due to duplicates")

	assert.NoError(t, errors[successIdx], "Succeeded creation should not return an error")
	assert.NotNil(t, results[successIdx], "Succeeded translation should not be nil")
	assert.Equal(t, "write", results[successIdx].EnglishWord, "EnglishWord should match")
	assert.Equal(t, "pisać", results[successIdx].PolishWord.Word, "PolishWord should match")
	assert.Equal(t, 1, len(results[successIdx].Examples), "Examples list should have 1 element")
	assert.Equal(t, "On lubi pisać listy.", results[successIdx].Examples[0].Sentence, "Sentence should match")

}

func TestConcurrentRemoveTranslation(t *testing.T) {

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
	id := createdTranslation.ID

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	start := false
	var wg sync.WaitGroup

	var results []bool
	var errors []error

	const iterations = 50
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			for !start {
				cond.Wait()
			}
			mu.Unlock()
			result, err := services.RemoveTranslation(db.GormTestDB, ctx, id)
			mu.Lock()
			results = append(results, result)
			errors = append(errors, err)
			mu.Unlock()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	start = true
	cond.Broadcast()
	mu.Unlock()

	wg.Wait()

	var successCount, errorCount int
	for i := 0; i < iterations; i++ {
		if results[i] {
			successCount++
		}
		if errors[i] != nil {
			errorCount++
		}
	}

	// Since further deletion doesn't change anything I decided that if at least one removal
	// is succesfull it is sufficient, because further deletions doesn't cause any problems
	assert.GreaterOrEqual(t, successCount, 1, "Expected at least one successful removal")
	assert.LessOrEqual(t, errorCount, iterations-1, "Expected the rest of the removals to fail due to non existent")

	_, err := services.Translation(db.GormTestDB, ctx, createdTranslation.ID)
	assert.Error(t, err, "Quering deleted translation should return an error")
}

// Test edge cases and errors
