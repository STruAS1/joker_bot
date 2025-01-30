package main

import (
	"SHUTKANULbot/bot"
	"SHUTKANULbot/config"
	"SHUTKANULbot/db"
)

func main() {
	cfg := config.LoadConfig()
	db.Connect(cfg)
	go bot.StartBot(cfg)
	select {}
}
