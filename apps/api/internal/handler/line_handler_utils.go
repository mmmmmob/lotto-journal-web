package handler

import (
	"strings"
	"unicode"

	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// sourceUserID extracts the LINE userId from a SourceInterface.
// Returns "" if the source is nil or is not a user source (e.g. group/room).
func sourceUserID(src webhook.SourceInterface) string {
	if src == nil {
		return ""
	}
	if us, ok := src.(webhook.UserSource); ok {
		return us.UserId
	}
	return ""
}

// Return if message sent is command for "List all tickets" or not.
func isTicketListCmd(text string) bool {
	normalized := normalizeCmd(text)
	return normalized == "โพย" || normalized == "list" || normalized == "tickets"
}

func isThaiSwitchCmd(text string) bool {
	normalized := normalizeCmd(text)
	return normalized == "ไทย" || normalized == "thai"
}

func isEnglishSwitchCmd(text string) bool {
	normalized := normalizeCmd(text)
	return normalized == "english" || normalized == "en"
}

func isAddHelpCmd(text string) bool {
	normalized := normalizeCmd(text)
	return normalized == "เพิ่ม" || normalized == "add"
}

func isNotifyHelpCmd(text string) bool {
	normalized := normalizeCmd(text)
	return normalized == "แจ้งเตือน" || normalized == "notify"
}

func normalizeCmd(text string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsSpace(r):
			return -1
		case r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\uFEFF':
			return -1
		default:
			return r
		}
	}, strings.ToLower(strings.TrimSpace(text)))
}

// webhookEventID extracts the webhookEventId from the event types we support.
// Returns "" for all other types — dispatch will log and skip those.
func webhookEventID(event webhook.EventInterface) string {
	switch e := event.(type) {
	case webhook.FollowEvent:
		return e.WebhookEventId
	case webhook.UnfollowEvent:
		return e.WebhookEventId
	case webhook.MessageEvent:
		return e.WebhookEventId
	case webhook.PostbackEvent:
		return e.WebhookEventId
	default:
		return ""
	}
}

func eventReplyToken(event webhook.EventInterface) string {
	switch e := event.(type) {
	case webhook.FollowEvent:
		return e.ReplyToken
	case webhook.MessageEvent:
		return e.ReplyToken
	case webhook.PostbackEvent:
		return e.ReplyToken
	default:
		return ""
	}
}

func ptr[T any](v T) *T {
	return &v
}

func parseQueryParams(data string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(data, "&")
	for _, part := range parts {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) == 2 {
			params[pair[0]] = pair[1]
		}
	}
	return params
}
