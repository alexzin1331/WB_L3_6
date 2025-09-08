package main

import (
	"log"

	"L3_6/internal/server"
	"L3_6/internal/storage"
	"L3_6/models"

	"github.com/ilyakaznacheev/cleanenv"
)

func loadConfig(path string) *models.Config {
	conf := &models.Config{}
	if err := cleanenv.ReadConfig(path, conf); err != nil {
		log.Fatal("Can't read the common config")
		return nil
	}
	return conf
}

func main() {
	cfg := loadConfig("config.yaml")

	db, err := storage.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	st := storage.NewStorage(db)
	srv := server.NewServer(st)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := srv.Run(cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
