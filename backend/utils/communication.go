package utils

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/glog"
	"net/http"
	"os"
	"strings"
)

// Load Discord webhook URL from environment variable
var discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")

func init() {
	glog.Info("Discord webhook address: %s", discordWebhookURL)
}

var capturers []func(string)

// AddDiscordSendCapturer capture or intercept discord sent notifications/messages for testing purposes
func AddDiscordSendCapturer(capturer func(string)) {
	capturers = append(capturers, capturer)
}

func SendToDiscord(content string) error {
	if capturers != nil {
		for _, capturer := range capturers {
			capturer(content)
		}
	}
	if discordWebhookURL == "" {
		_ = glog.Error("Discord webhook uls is not configured")
		return nil
	}

	payload := map[string]string{
		"content": content,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(discordWebhookURL, "application/json", strings.NewReader(string(payloadBytes)))
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send to Discord, status code: %d", resp.StatusCode)
	}

	return nil
}
