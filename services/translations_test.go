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

// Race condition occurs in this test since running it multiple times results in either result1 or
// result2 being nil (meaning that once result1 is nil and once result2 is nil)

// I could create a test running the following in a loop, checking if outcomes are different each time
// but since it is theoreticaly possible for the same result to be nil, I decided that this test is enough
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
	wg.Add(2)

	var result1, result2 *model.Translation
	var err1, err2 error

	go func() {
		defer wg.Done()
		mu.Lock()
		for !start {
			cond.Wait()
		}
		mu.Unlock()
		result1, err1 = services.CreateTranslation(db.GormTestDB, ctx, input)
	}()

	go func() {
		defer wg.Done()
		mu.Lock()
		for !start {
			cond.Wait()
		}
		mu.Unlock()
		result2, err2 = services.CreateTranslation(db.GormTestDB, ctx, input)
	}()

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	start = true
	cond.Broadcast()
	mu.Unlock()

	wg.Wait()

	if err1 == nil && err2 == nil {
		t.Fatalf("expected one creation to fail due to duplicate, but both succeeded")
	}
	if err1 != nil && err2 != nil {
		t.Fatalf("expected one creation to fail due to duplicate, but both failed")
	}

	var success *model.Translation
	var duplicateErr error
	if err1 == nil {
		success = result1
		duplicateErr = err2
	} else {
		success = result2
		duplicateErr = err1
	}

	assert.NotNil(t, success, "one translation creation should succeed: result1=%+v, result2=%+v", result1, result2)
	assert.Contains(t, duplicateErr.Error(), "already exists", fmt.Sprintf("expected duplicate error, got %v", duplicateErr))
	t.Logf("one translation creation should succeed: result1=%+v, result2=%+v", result1, result2)
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
	wg.Add(2)

	var result1, result2 bool
	var err1, err2 error

	go func() {
		defer wg.Done()
		mu.Lock()
		for !start {
			cond.Wait()
		}
		mu.Unlock()
		result1, err1 = services.RemoveTranslation(db.GormTestDB, ctx, id)
	}()

	go func() {
		defer wg.Done()
		mu.Lock()
		for !start {
			cond.Wait()
		}
		mu.Unlock()
		result2, err2 = services.RemoveTranslation(db.GormTestDB, ctx, id)
	}()

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	start = true
	cond.Broadcast()
	mu.Unlock()

	wg.Wait()

	if err1 == nil && err2 == nil {
		t.Fatalf("expected one removal to fail due to non existing id, but both succeeded")
	}
	if err1 != nil && err2 != nil {
		t.Fatalf("expected one removal to fail due to non existing id, but both failed")
	}

	var success bool
	var recordNotFound error
	if err1 == nil {
		success = result1
		recordNotFound = err2
	} else {
		success = result2
		recordNotFound = err1
	}

	assert.True(t, success, "one removal should succeed: result1=%+v, result2=%+v", result1, result2)
	assert.Contains(t, recordNotFound.Error(), "record not found", fmt.Sprintf("expected record not found error, got %v", recordNotFound))
	t.Logf("one translation removal should succeed: result1=%+v, result2=%+v", result1, result2)
}

// I decided that testing concurrent updates isn't necessary since the most recent one will simply
// override the previous update no matter which one is first, which is sufficient for this application

// Test edge cases and errors
