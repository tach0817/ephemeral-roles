// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

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
	FlightGroup             *singleflight.Group
}

// RoleNameFromChannel returns the name of a role for a channel, with the bot
// keyword prefixed.
func (handler *Handler) RoleNameFromChannel(channelName string) string {
	return fmt.Sprintf("%s %s", handler.RolePrefix, channelName)
}
