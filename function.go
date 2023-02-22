package codemagic_notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	sendMessageUrl   = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)
	telegramBotToken = os.Getenv("TELGRAM_BOT_TOKEN")
	telegramChatID   = os.Getenv("TELEGRAM_CHAT_ID")
	codeMagicAPIKey  = os.Getenv("CODEMAGIC_API_KEY")
)

func init() {
	functions.HTTP("codemagic_notifier", codeMagicNotifierFunction)
	fmt.Println("ENV VARS:")
	fmt.Println("TELEGRAM_BOT_TOKEN:", telegramBotToken)
	fmt.Println("TELEGRAM_CHAT_ID:", telegramChatID)
	fmt.Println("CODEMAGIC_API_KEY:", codeMagicAPIKey)
}

func codeMagicNotifierFunction(w http.ResponseWriter, r *http.Request) {
	var request []CodeMagicArtifactLink
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Panicln(fmt.Errorf("could not decode request body: %w", err))
		return
	}
	reqJson, err := json.Marshal(request)
	if err != nil {
		log.Println("request:", string(reqJson))
	}
	message := "<b>New build is available:</b>\n"

	for _, artifact := range request {
		if artifact.Type == "apk" {
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
			respBody, err := io.ReadAll(resp.Body)
			log.Println("public url response:", string(respBody))
			err = json.Unmarshal(respBody, &artifactPublicUrl)
			if err != nil {
				log.Println(fmt.Errorf("could not decode public url response: %w", err))
				continue
			}
			message += fmt.Sprintf(
				"%s (%s):\n<a href=\"%s\">[CodeMagic URL]</a>\n<a href=\"%s\">[Public URL]</a>\nExpires at: %s\n",
				artifact.Name,
				artifact.VersionName,
				artifact.Url,
				artifactPublicUrl.Url,
				artifactPublicUrl.ExpiresAt.String(),
			)
		}
	}
	chatID, err := strconv.ParseInt(telegramChatID, 10, 64)
	if err != nil {
		log.Println(fmt.Errorf("could not parse chat id: %w", err))
		return
	}
	sendMessageRequest := SendMessageRequest{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML",
	}

	req, err := json.Marshal(sendMessageRequest)
	if err != nil {
		log.Println(fmt.Errorf("could not marshal request: %w", err))
		return
	}
	resp, err := http.Post(sendMessageUrl, "application/json", bytes.NewReader(req))
	if err != nil {
		log.Println(fmt.Errorf("could not send request: %w", err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Println(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		b, err := io.ReadAll(resp.Body)
		if err == nil {
			log.Println(string(b))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
