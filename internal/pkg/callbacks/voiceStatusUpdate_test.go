package callbacks_test

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

func TestHandler_VoiceStateUpdate(t *testing.T) {
	jaegerTracer, jaegerCloser, err := tracer.New("test")
	if err != nil {
		t.Fatalf("Error creating Jaeger tracer: %s", err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			t.Errorf("Error closing Jaeger tracer: %s", err)
		}
	}()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	session.Client = http.NewClient(
		session.Client.Transport,
		jaegerTracer,
		"test-0",
	)

	defer mock.SessionClose(t, session)

	log := mock.NewLogger()

	monitorConfig := &monitor.Config{
		Log: log,
	}

	config := &callbacks.Handler{
		Log:                     log,
		BotName:                 "testBot",
		BotKeyword:              "testKeyword",
		RolePrefix:              "{eph}",
		JaegerTracer:            jaegerTracer,
		ContextTimeout:          time.Second,
		VoiceStateUpdateCounter: monitor.VoiceStateUpdateCounter(monitorConfig),
	}

	sendUpdate(session, config, mock.TestGuild, "unknownUser", mock.TestChannel)
	sendUpdate(session, config, mock.TestGuild, mock.TestUser, mock.TestPrivateChannel)
	sendUpdate(session, config, mock.TestGuild, mock.TestUser, mock.TestChannel)
	sendUpdate(session, config, mock.TestGuild, mock.TestUser, "")
	sendUpdate(session, config, mock.TestGuild, mock.TestUser, mock.TestChannel2)
	sendUpdate(session, config, mock.TestGuild, mock.TestUser, mock.TestChannel)
	sendUpdate(session, config, mock.TestGuild, mock.TestUser, "")
	sendUpdate(session, config, mock.TestGuildLarge, mock.TestUser, mock.TestChannel)
	sendUpdate(session, config, mock.TestGuildLarge, mock.TestUser, "")
}

func sendUpdate(session *discordgo.Session, config *callbacks.Handler, guildID, userID, channelID string) {
	config.VoiceStateUpdate(
		session,
		&discordgo.VoiceStateUpdate{
			VoiceState: &discordgo.VoiceState{
				UserID:    userID,
				GuildID:   guildID,
				ChannelID: channelID,
			},
		},
	)
}
