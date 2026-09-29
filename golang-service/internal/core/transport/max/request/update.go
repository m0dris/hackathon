package core_max_request

import (
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func SplitPayload(payload string) (key string, arg string) {
	i := strings.LastIndex(payload, ":")
	if i < 0 {
		return payload, ""
	}

	return payload[:i], payload[i+1:]
}

func AsMessageCreated(u schemes.UpdateInterface) (*schemes.MessageCreatedUpdate, bool) {
	v, ok := u.(*schemes.MessageCreatedUpdate)

	return v, ok
}

func AsMessageCallback(u schemes.UpdateInterface) (*schemes.MessageCallbackUpdate, bool) {
	v, ok := u.(*schemes.MessageCallbackUpdate)

	return v, ok
}

func AsBotStarted(u schemes.UpdateInterface) (*schemes.BotStartedUpdate, bool) {
	v, ok := u.(*schemes.BotStartedUpdate)

	return v, ok
}
