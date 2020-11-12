// +build gofuzz

package callbacks

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
	fuzz "github.com/google/gofuzz"
)

type testingInstance struct{}

func (ti *testingInstance) Error(args ...interface{}) {

}

func Fuzz(data []byte) int {
	jaegerTracer, jaegerCloser, err := tracer.New("test")
	if err != nil {
		panic(err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			panic(err)
		}
	}()

	session, err := mock.NewSession()
	if err != nil {
		panic(err)
	}

	session.Client = http.NewClient(
		session.Client.Transport,
		jaegerTracer,
		"test-0",
	)

	defer mock.SessionClose(&testingInstance{}, session)

	log := mock.NewLogger()

	monitorConfig := &monitor.Config{
		Log: log,
	}

	config := &Config{
		Log:                     log,953u
		BotName:                 "testBot",
		BotKeyword:              "testKeyword",
		RolePrefix:              "{eph}",
		JaegerTracer:            jaegerTracer,
		ContextTimeout:          time.Second,
		VoiceStateUpdateCounter: monitor.VoiceStateUpdateCounter(monitorConfig),
	}

	voiceStateUpdate := &discordgo.VoiceStateUpdate{}

	fuzz.NewFromGoFuzz(data).Fuzz(voiceStateUpdate)

	config.VoiceStateUpdate(session, voiceStateUpdate)

	return 0
}
