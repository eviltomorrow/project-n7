package conf

import (
	"bytes"
	"encoding/json"
	"os"

	jsoniter "github.com/json-iterator/go"
)

type TelegramBot struct {
	Token string `json:"token"`

	Domain  string `json:"domain"`
	WebHook string `json:"web-hook"`
}

func (t *TelegramBot) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(t)
	return string(buf)
}

func FindTelegramBot(path string) (*TelegramBot, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data = bytes.TrimSpace(buf)
	var s = &TelegramBot{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}
