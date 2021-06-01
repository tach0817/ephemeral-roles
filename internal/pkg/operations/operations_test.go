package operations_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	"golang.org/x/sync/singleflight"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const newRoleID = "newRole"

type sessionFunc func() *discordgo.Session

type roleForMemberTestCase struct {
	name       string
	guildID    string
	userID     string
	roleID     string
	getSession sessionFunc
	testFunc   func(
		t *testing.T,
		getSession sessionFunc,
		guildID, userID, roleName string,
	)
}

func TestLookupGuild(t *testing.T) {
	flightGroup := &singleflight.Group{}

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	_, err = operations.LookupGuild(flightGroup, session, mockconstants.TestGuild)
	if err != nil {
		t.Fatal(err)
	}

	_, err = operations.LookupGuild(flightGroup, session, mockconstants.TestGuildLarge)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddRoleToMember(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	getSession := func() *discordgo.Session { return session }

	runRoleForMemberTestCases(t, addRoleToMemberTestCases(getSession))
}

func TestRemoveRoleFromMember(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	getSession := func() *discordgo.Session { return session }

	runRoleForMemberTestCases(t, removeRoleFromMemberTestCases(getSession))
}

func TestIsDeadlineExceeded(t *testing.T) {
	if operations.IsDeadlineExceeded(io.EOF) {
		t.Errorf("Unexpected success")
	}

	if !operations.IsDeadlineExceeded(&callbacks.DeadlineExceeded{Err: context.DeadlineExceeded}) {
		t.Errorf("Unexpected failure")
	}
}

func TestIsForbiddenResponse(t *testing.T) {
	type testCase struct {
		name     string
		expected bool
		err      error
	}

	testCases := []*testCase{
		{
			name:     "nil error",
			expected: false,
			err:      nil,
		},
		{
			name:     "non-nil error",
			expected: false,
			err:      io.EOF,
		},
		{
			name:     "*discordgo.RESTError http.StatusInternalServerError",
			expected: false,
			err:      &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusInternalServerError}},
		},
		{
			name:     "*discordgo.RESTError http.StatusForbidden",
			expected: true,
			err:      &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusForbidden}},
		},
		{
			name:     "wrapped *discordgo.RESTError http.StatusForbidden",
			expected: true,
			err:      fmt.Errorf("%w", &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusForbidden}}),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			isForbiddenResponse(t, testCase.expected, testCase.err)
		})
	}
}

func TestIsMaxGuildsResponse(t *testing.T) {
	if operations.IsMaxGuildsResponse(io.EOF) {
		t.Errorf("Unexpected success")
	}

	maxGuildsResponse := &discordgo.RESTError{
		Response: &http.Response{StatusCode: http.StatusBadRequest},
		Message:  &discordgo.APIErrorMessage{Code: operations.APIErrorCodeMaxRoles},
	}

	if !operations.IsMaxGuildsResponse(maxGuildsResponse) {
		t.Errorf("Unexpected failure")
	}
}

func TestShouldLogDebug(t *testing.T) {
	if operations.ShouldLogDebug(io.EOF) {
		t.Errorf("Unexpected success")
	}

	if !operations.ShouldLogDebug(&callbacks.DeadlineExceeded{Err: context.DeadlineExceeded}) {
		t.Errorf("Unexpected failure")
	}
}

func TestBotHasChannelPermission(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	testChannelWithPermission, err := session.State.Channel(mockconstants.TestChannel)
	if err != nil {
		t.Fatal(err)
	}

	testChannelWithoutPermission, err := session.State.Channel(mockconstants.TestPrivateChannel)
	if err != nil {
		t.Fatal(err)
	}

	err = operations.BotHasChannelPermission(session, testChannelWithPermission)
	if err != nil {
		t.Error(err)
	}

	err = operations.BotHasChannelPermission(session, testChannelWithoutPermission)
	if err == nil {
		t.Error("unexpected nil error")
	}
}

func addRoleToMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "add role user does not have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
		{
			name:       "add role user does have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
	}
}

func removeRoleFromMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "remove role member does have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
		{
			name:       "remove role member does not have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
	}
}

func addNewRoleToMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	roleForMember(t, getSession, guildID, userID, roleID, true)
}

func removeRoleFromMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	roleForMember(t, getSession, guildID, userID, roleID, false)
}

func roleForMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
	add bool,
) {
	session := getSession()

	switch add {
	case true:
		err := operations.AddRoleToMember(session, guildID, userID, roleID)
		if err != nil {
			t.Errorf("unexpected error adding role to member: %s", err)
		}
	case false:
		err := operations.RemoveRoleFromMember(session, guildID, userID, roleID)
		if err != nil {
			t.Errorf("unexpected error removing role from member: %s", err)
		}
	}
}

func runRoleForMemberTestCases(t *testing.T, testCases []*roleForMemberTestCase) {
	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			testCase.testFunc(t, testCase.getSession, testCase.guildID, testCase.userID, testCase.roleID)
		})
	}
}

func isForbiddenResponse(t *testing.T, expected bool, err error) {
	actual := operations.IsForbiddenResponse(err)

	if actual != expected {
		t.Errorf("unexpected forbidden response: %t", actual)
	}
}
