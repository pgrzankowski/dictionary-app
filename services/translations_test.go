// services/translation_test.go
package services_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pgrzankowski/dictionary-app/db"
	"github.com/pgrzankowski/dictionary-app/graph/model"
	"github.com/pgrzankowski/dictionary-app/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupMockDB creates a new sqlmock connection and returns a GORM DB instance along with the mock.
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

func TestCreateTranslation(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	db.GormDB = gormDB

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

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "translations" ("polish_word_id","english_word","created_at","updated_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)).
		WithArgs(1, input.EnglishWord, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "examples" ("translation_id","sentence","created_at","updated_at") VALUES ($1,$2,$3,$4) RETURNING "id"`)).
		WithArgs(1, input.Examples[0].Sentence, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	ctx := context.Background()
	translation, err := services.CreateTranslation(ctx, input)
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
