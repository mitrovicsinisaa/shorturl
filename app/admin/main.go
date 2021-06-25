package main

import (
	"fmt"
	"log"

	"github.com/mitrovicsinisaa/shorturl/business/data/schema"
	"github.com/mitrovicsinisaa/shorturl/foundation/database"
)

func main() {
	migrate()
}
func migrate() {
	cfg := database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       "0.0.0.0",
		Name:       "postgres",
		DisableTLS: true,
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("migrations complete")

	if err := schema.Seed(db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seed data complete")
}
