package utils

import (
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

func SendToDiscord(payload []byte) error {
	resp, err := http.Post(discordWebhookURL, "application/json", strings.NewReader(string(payload)))
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
