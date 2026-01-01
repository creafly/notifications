package i18n

import (
	"encoding/json"
	"os"
	"sync"
)

type Locale string

const (
	LocaleEnUS Locale = "en-US"
	LocaleRuRU Locale = "ru-RU"
)

type Messages struct {
	Errors       ErrorMessages        `json:"errors"`
	Notification NotificationMessages `json:"notification"`
	Invitation   InvitationMessages   `json:"invitation"`
}

type ErrorMessages struct {
	InternalError    string `json:"internalError"`
	ValidationFailed string `json:"validationFailed"`
	Unauthorized     string `json:"unauthorized"`
	NotFound         string `json:"notFound"`
}

type NotificationMessages struct {
	MarkedRead    string `json:"markedRead"`
	AllMarkedRead string `json:"allMarkedRead"`
	Deleted       string `json:"deleted"`
}

type InvitationMessages struct {
	Created  string `json:"created"`
	Accepted string `json:"accepted"`
	Rejected string `json:"rejected"`
	Expired  string `json:"expired"`
	NotFound string `json:"notFound"`
}

var (
	messagesCache = make(map[Locale]*Messages)
	cacheMutex    sync.RWMutex
)

func GetMessages(locale Locale) *Messages {
	cacheMutex.RLock()
	if messages, ok := messagesCache[locale]; ok {
		cacheMutex.RUnlock()
		return messages
	}
	cacheMutex.RUnlock()

	messages := loadMessages(locale)
	if messages == nil {
		messages = loadMessages(LocaleEnUS)
	}

	cacheMutex.Lock()
	messagesCache[locale] = messages
	cacheMutex.Unlock()

	return messages
}

func loadMessages(locale Locale) *Messages {
	filename := "resources/" + string(locale) + ".json"
	data, err := os.ReadFile(filename)
	if err != nil {
		return getDefaultMessages()
	}

	var messages Messages
	if err := json.Unmarshal(data, &messages); err != nil {
		return getDefaultMessages()
	}

	return &messages
}

func getDefaultMessages() *Messages {
	return &Messages{
		Errors: ErrorMessages{
			InternalError:    "Internal server error",
			ValidationFailed: "Validation failed",
			Unauthorized:     "Unauthorized",
			NotFound:         "Not found",
		},
		Notification: NotificationMessages{
			MarkedRead:    "Notification marked as read",
			AllMarkedRead: "All notifications marked as read",
			Deleted:       "Notification deleted",
		},
		Invitation: InvitationMessages{
			Created:  "Invitation sent",
			Accepted: "Invitation accepted",
			Rejected: "Invitation rejected",
			Expired:  "Invitation expired",
			NotFound: "Invitation not found",
		},
	}
}

func PreloadLocales() {
	GetMessages(LocaleEnUS)
	GetMessages(LocaleRuRU)
}
