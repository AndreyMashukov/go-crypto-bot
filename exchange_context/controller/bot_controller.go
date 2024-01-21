package controller

import (
	"encoding/json"
	"fmt"
	"github.com/AndreyMashukov/go-crypto-bot/server/exchange_context/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/exchange_context/service"
	"net/http"
)

type BotController struct {
	HealthService *service.HealthService
	CurrentBot    *model.Bot
}

func (b *BotController) GetHealthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != b.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	health := b.HealthService.HealthCheck()

	encoded, _ := json.Marshal(health)
	fmt.Fprintf(w, string(encoded))
}
