package codemagic_notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

var (
	sendMessageUrl   = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)
	telegramBotToken = os.Getenv("TELGRAM_BOT_TOKEN")
	telegramChatID   = os.Getenv("TELEGRAM_CHAT_ID")
	codeMagicAPIKey  = os.Getenv("CODEMAGIC_API_KEY")
)

func init() {
	functions.HTTP("codemagic_notifier", codeMagicNotifierFunction)
}

func codeMagicNotifierFunction(w http.ResponseWriter, r *http.Request) {
	var request []CodeMagicArtifactLink
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Panicln(fmt.Errorf("could not decode request body: %w", err))
		return
	}

	message := "New build is available:\n"

	for _, artifact := range request {
		body := strings.NewReader(
			fmt.Sprintf(
				`{"expiresAt": %d}`,
				time.Now().Add(time.Hour*24*31*6).Unix(),
			),
		)
		req, err := http.NewRequest("POST", artifact.Url+"/public-url", body)
		req.Header.Add("x-auth-token", codeMagicAPIKey)
		req.Header.Add("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(fmt.Errorf("could not get public url: %w", err))
			continue
		}
		var artifactPublicUrl CodeMagicArtifactPublicUrl
		err = json.NewDecoder(resp.Body).Decode(&artifact)
		if err != nil {
			log.Println(fmt.Errorf("could not decode public url response: %w", err))
			continue
		}
		message += fmt.Sprintf(
			"%s (%s): %s\nExpires at:%s\n",
			artifact.Name,
			artifact.VersionName,
			artifactPublicUrl.Url,
			artifactPublicUrl.ExpiresAt.Format("01-02-2006 15:04:05 Mon"),
		)
	}
	chatID, err := strconv.ParseInt(telegramChatID, 10, 64)
	if err != nil {
		log.Println(fmt.Errorf("could not parse chat id: %w", err))
		return
	}
	sendMessageRequest := SendMessageRequest{
		ChatID: chatID,
		Text:   message,
	}

	req, err := json.Marshal(sendMessageRequest)
	if err != nil {
		log.Println(fmt.Errorf("could not marshal request: %w", err))
		return
	}
	_, err = http.Post(sendMessageUrl, "application/json", bytes.NewReader(req))
	if err != nil {
		log.Println(fmt.Errorf("could not send request: %w", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
