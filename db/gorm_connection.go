package db

import (
	"fmt"
	"log"
	"os"

	"github.com/pgrzankowski/dictionary-app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GormDB *gorm.DB
var GormTestDB *gorm.DB

func ConnectGORM() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	var err error
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to GORM database: %v", err)
	}

	if err := GormDB.AutoMigrate(&models.PolishWord{}); err != nil {
		log.Fatalf("AutoMigrate PolishWord failed: %v", err)
	}
	if err := GormDB.AutoMigrate(&models.Translation{}); err != nil {
		log.Fatalf("AutoMigrate Translation failed: %v", err)
	}
	if err := GormDB.AutoMigrate(&models.Example{}); err != nil {
		log.Fatalf("AutoMigrate Example failed: %v", err)
	}

	log.Printf("Connected to database using GORM: %s", dsn)
}

func ConnectTestGORM() {
	host := os.Getenv("DB_TEST_HOST")
	port := os.Getenv("DB_TEST_PORT")
	user := os.Getenv("DB_TEST_USER")
	password := os.Getenv("DB_TEST_PASS")
	dbname := os.Getenv("DB_TEST_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	GormTestDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to GORM database: %v", err)
	}

	if err := GormTestDB.AutoMigrate(&models.PolishWord{}); err != nil {
		log.Fatalf("AutoMigrate PolishWord failed: %v", err)
	}
	if err := GormTestDB.AutoMigrate(&models.Translation{}); err != nil {
		log.Fatalf("AutoMigrate Translation failed: %v", err)
	}
	if err := GormTestDB.AutoMigrate(&models.Example{}); err != nil {
		log.Fatalf("AutoMigrate Example failed: %v", err)
	}

	log.Printf("Connected to database using GORM: %s", dsn)
}
