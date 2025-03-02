package services_test

import (
	"context"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pgrzankowski/dictionary-app/graph/model"
	"github.com/pgrzankowski/dictionary-app/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbConn,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	return gormDB, mock
}

// Test mutations
func TestCreateTranslation(t *testing.T) {
	gormDB, mock := setupMockDB(t)

	input := model.NewTranslationInput{
		PolishWord:  "pisać",
		EnglishWord: "write",
		Examples: []*model.NewExampleInput{
			{Sentence: "On lubi pisać listy."},
		},
	}

	mock.ExpectBegin()

	mock.ExpectQuery(`(?i)^SELECT \* FROM "polish_words".*LIMIT \$3`).
		WithArgs(input.PolishWord, input.PolishWord, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
			AddRow(1, input.PolishWord, time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE polish_word_id = $1 AND english_word = $2 ORDER BY "translations"."id" LIMIT $3`)).
		WithArgs(1, input.EnglishWord, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}))

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "translations" ("polish_word_id","english_word","created_at","updated_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)).
		WithArgs(1, input.EnglishWord, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "examples" ("translation_id","sentence","created_at","updated_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)).
		WithArgs(1, input.Examples[0].Sentence, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	ctx := context.Background()
	translation, err := services.CreateTranslation(gormDB, ctx, input)
	if err != nil {
		t.Fatalf("CreateTranslation failed: %v", err)
	}
	if translation.ID != "1" || translation.EnglishWord != input.EnglishWord {
		t.Errorf("unexpected translation returned: %+v", translation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestRemoveTranslation(t *testing.T) {
	gormDB, mock := setupMockDB(t)

	translationID := 1
	polishWordID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`(?i)^SELECT \* FROM "translations" WHERE "translations"."id" = \$1 ORDER BY "translations"."id" LIMIT \$2`).
		WithArgs(translationID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}).
			AddRow(translationID, polishWordID, "write", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
		WithArgs(polishWordID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
			AddRow(polishWordID, "pisać", time.Now(), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "translations" WHERE "translations"."id" = $1`)).
		WithArgs(translationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM "translations" WHERE polish_word_id = $1`)).
		WithArgs(polishWordID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "polish_words" WHERE "polish_words"."id" = $1`)).
		WithArgs(polishWordID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	result, err := services.RemoveTranslation(gormDB, context.Background(), "1")
	if err != nil {
		t.Fatalf("RemoveTranslation failed: %v", err)
	}
	if !result {
		t.Error("expected removal to succeed, got false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateTranslation(t *testing.T) {
	gormDB, mock := setupMockDB(t)

	newEng := "modify"
	input := model.UpdateTranslationInput{
		ID:          "1",
		EnglishWord: &newEng,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE "translations"."id" = $1 ORDER BY "translations"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "polish_word_id", "english_word", "created_at", "updated_at",
		}).AddRow(1, 1, "write", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "examples" WHERE "examples"."translation_id" = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "translation_id", "sentence", "created_at", "updated_at",
		}))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
			AddRow(1, "pisać", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "polish_words" ("word","created_at","updated_at","id") VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING RETURNING "id"`)).
		WithArgs("pisać", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "translations" SET "polish_word_id"=$1,"english_word"=$2,"created_at"=$3,"updated_at"=$4 WHERE "id" = $5`)).
		WithArgs(1, "modify", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	ctx := context.Background()
	updated, err := services.UpdateTranslation(gormDB, ctx, input)
	if err != nil {
		t.Fatalf("UpdateTranslation failed: %v", err)
	}

	if updated.EnglishWord != "modify" {
		t.Errorf("expected english word to be 'modify', got %s", updated.EnglishWord)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// Test queries
func TestTranslations(t *testing.T) {
	gormDB, mock := setupMockDB(t)

	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}).
			AddRow(1, 1, "write", now, now))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "examples" WHERE "examples"."translation_id" = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "translation_id", "sentence", "created_at", "updated_at"}))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
			AddRow(1, "pisać", now, now))

	ctx := context.Background()
	result, err := services.Translations(gormDB, ctx)
	if err != nil {
		t.Fatalf("Translations query failed: %v", err)
	}

	assert.Len(t, result, 1, "expected 1 translation")

	translation := result[0]
	assert.Equal(t, "1", translation.ID, "expected translation ID to be '1'")
	assert.Equal(t, "write", translation.EnglishWord, "expected english word to be 'write'")
	if translation.PolishWord == nil {
		t.Error("expected a non-nil PolishWord")
	} else {
		assert.Equal(t, "pisać", translation.PolishWord.Word, "expected polish word to be 'pisać'")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestTranslationByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)

	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE "translations"."id" = $1 ORDER BY "translations"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}).
			AddRow(1, 1, "write", now, now))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "examples" WHERE "examples"."translation_id" = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "translation_id", "sentence", "created_at", "updated_at"}))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
			AddRow(1, "pisać", now, now))

	ctx := context.Background()
	result, err := services.Translation(gormDB, ctx, "1")
	if err != nil {
		t.Fatalf("Translation query failed: %v", err)
	}

	assert.NotNil(t, result, "expected translation")
	assert.Equal(t, "1", result.ID, "expected translation ID to be '1'")
	assert.Equal(t, "write", result.EnglishWord, "expected english word to be 'write'")
	if result.PolishWord == nil {
		t.Error("expected PolishWord")
	} else {
		assert.Equal(t, "pisać", result.PolishWord.Word, "expected polish word to be 'pisać'")
	}
	assert.Len(t, result.Examples, 0, "expected 0 examples")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test concurrency
func TestConcurrentCreateTranslation(t *testing.T) {
	const concurrency = 5
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(instance int) {
			defer wg.Done()
			gormDB, mock := setupMockDB(t)

			input := model.NewTranslationInput{
				PolishWord:  "pisać",
				EnglishWord: "write",
				Examples: []*model.NewExampleInput{
					{Sentence: "On lubi pisać listy."},
				},
			}

			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE word = $1 AND "polish_words"."word" = $2 ORDER BY "polish_words"."id" LIMIT $3`)).
				WithArgs(input.PolishWord, input.PolishWord, 1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}))

			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "polish_words" ("word","created_at","updated_at") VALUES ($1,$2,$3) RETURNING "id"`)).
				WithArgs(input.PolishWord, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE polish_word_id = $1 AND english_word = $2 ORDER BY "translations"."id" LIMIT $3`)).
				WithArgs(1, input.EnglishWord, 1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}))

			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "translations" ("polish_word_id","english_word","created_at","updated_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)).
				WithArgs(1, input.EnglishWord, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "examples" ("translation_id","sentence","created_at","updated_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)).
				WithArgs(1, input.Examples[0].Sentence, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			mock.ExpectCommit()

			ctx := context.Background()
			tr, err := services.CreateTranslation(gormDB, ctx, input)
			if err != nil {
				t.Errorf("instance %d: CreateTranslation failed: %v", instance, err)
			}
			if tr == nil || tr.ID != "1" {
				t.Errorf("instance %d: unexpected translation result: %+v", instance, tr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("instance %d: unfulfilled expectations: %v", instance, err)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentUpdateTranslation(t *testing.T) {
	const concurrency = 5
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(instance int) {
			defer wg.Done()
			gormDB, mock := setupMockDB(t)

			newEng := "modify"
			input := model.UpdateTranslationInput{
				ID:          "1",
				EnglishWord: &newEng,
			}

			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE "translations"."id" = $1 ORDER BY "translations"."id" LIMIT $2`)).
				WithArgs(1, 1).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "polish_word_id", "english_word", "created_at", "updated_at",
				}).AddRow(1, 1, "write", time.Now(), time.Now()))

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "examples" WHERE "examples"."translation_id" = $1`)).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "translation_id", "sentence", "created_at", "updated_at"}))

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
					AddRow(1, "pisać", time.Now(), time.Now()))

			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "polish_words" ("word","created_at","updated_at","id") VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING RETURNING "id"`)).
				WithArgs("pisać", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			mock.ExpectExec(regexp.QuoteMeta(`UPDATE "translations" SET "polish_word_id"=$1,"english_word"=$2,"created_at"=$3,"updated_at"=$4 WHERE "id" = $5`)).
				WithArgs(1, "modify", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			ctx := context.Background()
			updated, err := services.UpdateTranslation(gormDB, ctx, input)
			if err != nil {
				t.Errorf("instance %d: UpdateTranslation failed: %v", instance, err)
			}
			if updated == nil || updated.EnglishWord != "modify" {
				t.Errorf("instance %d: unexpected update result: %+v", instance, updated)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("instance %d: unfulfilled expectations: %v", instance, err)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentRemoveTranslation(t *testing.T) {
	const concurrency = 5
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(instance int) {
			defer wg.Done()
			gormDB, mock := setupMockDB(t)

			translationID := 1
			polishWordID := 1

			mock.ExpectBegin()

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE "translations"."id" = $1 ORDER BY "translations"."id" LIMIT $2`)).
				WithArgs(translationID, 1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}).
					AddRow(translationID, polishWordID, "write", time.Now(), time.Now()))

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
				WithArgs(polishWordID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
					AddRow(polishWordID, "pisać", time.Now(), time.Now()))

			mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "translations" WHERE "translations"."id" = $1`)).
				WithArgs(translationID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "translations" WHERE polish_word_id = $1`)).
				WithArgs(polishWordID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "polish_words" WHERE "polish_words"."id" = $1`)).
				WithArgs(polishWordID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectCommit()

			ctx := context.Background()
			res, err := services.RemoveTranslation(gormDB, ctx, strconv.Itoa(translationID))
			if err != nil {
				t.Errorf("instance %d: RemoveTranslation failed: %v", instance, err)
			}
			if res != true {
				t.Errorf("instance %d: expected removal to succeed", instance)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("instance %d: unfulfilled expectations: %v", instance, err)
			}
		}(i)
	}
	wg.Wait()
}

// Edge cases
func TestRemoveTranslation_PolishWordNotRemoved(t *testing.T) {
	gormDB, mock := setupMockDB(t)

	translationID := 1
	polishWordID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "translations" WHERE "translations"."id" = $1 ORDER BY "translations"."id" LIMIT $2`)).
		WithArgs(translationID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "polish_word_id", "english_word", "created_at", "updated_at"}).
			AddRow(translationID, polishWordID, "write", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "polish_words" WHERE "polish_words"."id" = $1`)).
		WithArgs(polishWordID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "created_at", "updated_at"}).
			AddRow(polishWordID, "pisać", time.Now(), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "translations" WHERE "translations"."id" = $1`)).
		WithArgs(translationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "translations" WHERE polish_word_id = $1`)).
		WithArgs(polishWordID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectCommit()

	ctx := context.Background()
	result, err := services.RemoveTranslation(gormDB, ctx, strconv.Itoa(translationID))
	if err != nil {
		t.Fatalf("RemoveTranslation failed: %v", err)
	}
	assert.True(t, result, "expected removal to succeed and return true")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
