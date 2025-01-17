// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

// OperationsGateway is an interface abstraction for processing operations
// requests.
type OperationsGateway interface {
	Process(operations.ResultChannel, *operations.Request)
}

// Handler contains fields for the callback methods attached to it.
type Handler struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	RoleColor               int
	JaegerTracer            opentracing.Tracer
	ContextTimeout          time.Duration
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
	OperationsGateway       OperationsGateway
}

// RoleNameFromChannel returns the name of a role for a channel, with the bot
// keyword prefixed.
func (handler *Handler) RoleNameFromChannel(channelName string) string {
	return fmt.Sprintf("%s %s", handler.RolePrefix, channelName)
}
