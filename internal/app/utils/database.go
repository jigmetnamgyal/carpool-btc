package utils

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var DB *sql.DB

func ConnectToDb() {
	var connectionStr string

	log.Println(os.Getenv("ENVIRONMENT"))

	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		connectionStr = os.Getenv("DATABASE_URL")
	} else {
		connectionStr = os.Getenv("TEST_DATABASE_URL")
	}

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		log.Fatal("Error connecting to db")
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to create db connection:" + err.Error())
	}

	DB = db
}
