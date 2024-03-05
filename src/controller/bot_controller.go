package controller

import (
	"encoding/json"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"net/http"
)

type BotController struct {
	HealthService *service.HealthService
	CurrentBot    *model.Bot
	BotRepository *repository.BotRepository
}

func (b *BotController) GetHealthCheckAction(w http.ResponseWriter, req *http.Request) {
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

func (b *BotController) PutConfigAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != b.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	if req.Method != "PUT" {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)

		return
	}

	var botUpdate model.BotConfigUpdate

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&botUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	bot := b.BotRepository.GetCurrentBot()
	bot.IsMasterBot = botUpdate.IsMasterBot
	bot.IsSwapEnabled = botUpdate.IsSwapEnabled
	err = b.BotRepository.Update(*bot)

	if err != nil {
		http.Error(w, "Couldn't update bot config.", http.StatusBadRequest)

		return
	}
}
