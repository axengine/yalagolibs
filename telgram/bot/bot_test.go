package bot

import (
	"context"
	"testing"
)

func TestSendTelegramMessage(t *testing.T) {
	botToken := "7703216837:AAEcmP9rhx-cBORm0vAG6SwRyEYqpszrlUs"
	chatID := "7958836571"
	message := "Server monitoring alert: CPU usage exceeds 90%!"

	if err := SendTelegramMessage(context.Background(), botToken, chatID, message); err != nil {
		t.Fatal(err)
	} else {
		t.Log("send success！")
	}
}
