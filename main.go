package main

import (
	TonConnectCallback "SHUTKANULbot/TonConnectCallBack"
	"SHUTKANULbot/bot"
	"SHUTKANULbot/config"
	"SHUTKANULbot/db"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func main() {
	cfg := config.LoadConfig()
	db.Connect(cfg)
	go bot.StartBot(cfg)
	go StartTonConnectServer()
	client := liteclient.NewConnectionPool()
	cfgs, err := liteclient.GetConfigFromUrl(context.Background(), "https://ton.org/global-config.json")
	if err != nil {
		log.Fatalln("get config err: ", err.Error())
		return
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfgs)
	if err != nil {
		log.Fatalln("connection err: ", err.Error())
		return
	}
	api := ton.NewAPIClient(client)
	ctx := client.StickyContext(context.Background())

	minterAddr := "0:42a3dab99606812e24cf919c056757656769791a0efa6d3e7f7939a5d1fcd9c9"
	ownerWallet := "0:d3932b6b42c4fe4b4f0f84f9fe5a3833710cfed8107027c14b139f0533df44db"

	adminWallet, err := wallet.FromSeed(api, strings.Split(cfg.FromSeed, " "), wallet.V4R2)
	if err != nil {
		log.Fatal("Ошибка загрузки кошелька админа:", err)
	}
	go TonConnectCallback.TransactionWorker(db.DB, api, ctx, adminWallet, minterAddr, ownerWallet)
	select {}
}

var (
	imageData  []byte
	imageMutex sync.Mutex
	loaded     bool
)

func StartTonConnectServer() {
	http.HandleFunc("/tonconnect-manifest.json", handleManifest)
	http.HandleFunc("/pic", handleImageFromMemory)

	log.Println("🚀 TON Connect HTTPS-сервер запущен на порту 443...")
	err := http.ListenAndServeTLS(":8999", "certificate.crt", "certificate.key", nil)
	if err != nil {
		log.Fatalf("Ошибка при запуске HTTPS-сервера: %v", err)
	}
}

func handleManifest(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfig()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	manifest := map[string]interface{}{
		"name":    "JOKER",
		"url":     "https://t.me/gasgagasgagagabot",
		"iconUrl": fmt.Sprintf("https://%s/pic", cfg.Domines),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manifest)
}

func handleImageFromMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Только GET, чувак!", http.StatusMethodNotAllowed)
		return
	}
	imageMutex.Lock()
	defer imageMutex.Unlock()

	if !loaded {
		data, err := os.ReadFile("photos/test.png")
		if err != nil {
			log.Printf("Не смогли прочитать файл: %v", err)
			http.Error(w, "Ошибка при чтении файла", http.StatusInternalServerError)
			return
		}

		imageData = data
		loaded = true
		log.Println("Картинка загружена в память. Больше в файл не лезем.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(imageData)
}
