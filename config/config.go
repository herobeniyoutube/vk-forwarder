package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	VkGroupToken       string
	VkGroupID          string
	VkVersion          string
	TgBotToken         string
	TgChatID           string
	Environment        string
	DBConnectionString string
	GoEnv              string
}

func NewConfig() Config {
	cfg := Config{
		VkGroupToken: os.Getenv("vk_group_token"),
		VkGroupID:    os.Getenv("vk_group_id"),
		VkVersion:    os.Getenv("vk_version"),
		TgBotToken:   os.Getenv("tg_bot_token"),
		TgChatID:     os.Getenv("tg_chat_id"),
		Environment:  os.Getenv("environment"),
		GoEnv:        os.Getenv("go_env"),
	}

	dbUri := cfg.buildPostgresUrl()
	if cfg.Environment == "dev" {
		log.Println(dbUri)
	}
	cfg.DBConnectionString = dbUri

	return cfg
}

func (c *Config) buildPostgresUrl() string {
	sslMode := "disable"

	dbUri := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("db_host"),
		os.Getenv("db_port"),
		os.Getenv("db_user"),
		os.Getenv("db_name"),
		os.Getenv("db_pass"),
		sslMode,
	)

	if c.GoEnv != "prod" {
		fmt.Println(dbUri)
	}

	return dbUri
}
