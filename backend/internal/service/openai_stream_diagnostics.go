package service

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	openAIStreamInterruptionKindHTTPPassthroughMissingTerminal = "stream_interrupted:http_passthrough_missing_terminal"
	openAIStreamInterruptionKindHTTPPassthroughReadError       = "stream_interrupted:http_passthrough_read_error"
	openAIStreamInterruptionKindHTTPResponsesMissingTerminal   = "stream_interrupted:http_responses_missing_terminal"
	openAIStreamInterruptionKindHTTPResponsesReadError         = "stream_interrupted:http_responses_read_error"
	openAIStreamInterruptionKindHTTPResponsesTimeout           = "stream_interrupted:http_responses_timeout"
	openAIStreamInterruptionKindWSAfterDownstreamWrite         = "stream_interrupted:ws_after_downstream_write"
)

func formatOpenAIStreamInterruptionDetail(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, " ")
}

func recordOpenAIStreamInterruption(
	c *gin.Context,
	account *Account,
	requestID string,
	kind string,
	message string,
	passthrough bool,
	detail string,
) {
	if c == nil {
		return
	}
	message = strings.TrimSpace(message)
	detail = strings.TrimSpace(detail)
	if message == "" {
		return
	}

	setOpsUpstreamError(c, 0, message, detail)

	ev := OpsUpstreamErrorEvent{
		Passthrough:       passthrough,
		Platform:          PlatformOpenAI,
		UpstreamRequestID: strings.TrimSpace(requestID),
		Kind:              strings.TrimSpace(kind),
		Message:           message,
		Detail:            detail,
	}
	if account != nil {
		ev.AccountID = account.ID
		ev.AccountName = account.Name
		if strings.TrimSpace(account.Platform) != "" {
			ev.Platform = account.Platform
		}
	}

	appendOpsUpstreamError(c, ev)
}
