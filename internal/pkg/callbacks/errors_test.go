package callbacks_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	wrapMsg               = "wrapped error"
	invalidErrorAssertion = "Invalid error assertion"
)

func TestRoleNotFound_Error(t *testing.T) {
	rnf := &callbacks.RoleNotFound{}

	if rnf.Error() == "" {
		t.Error("unexpected empty error message")
	}
}

func TestMemberNotFound_Is(t *testing.T) {
	mnf := &callbacks.MemberNotFound{}

	if errors.Is(nil, &callbacks.MemberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.MemberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnf, &callbacks.MemberNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestMemberNotFound_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	mnf := &callbacks.MemberNotFound{Err: wrappedErr}

	unwrappedErr := mnf.Unwrap()

	if unwrappedErr != wrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestMemberNotFound_Error(t *testing.T) {
	mnf := &callbacks.MemberNotFound{}
	expectedErrMsg := callbacks.MemberNotFoundMessage

	if mnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnf.Error(),
			expectedErrMsg,
		)
	}

	mnf.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if mnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnf.Error(),
			expectedErrMsg,
		)
	}
}

func TestMemberNotFound_Guild(t *testing.T) {
	expected := &discordgo.Guild{Name: mock.TestGuild}
	mnf := &callbacks.MemberNotFound{Guild: expected}
	actual := mnf.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_Member(t *testing.T) {
	var expected *discordgo.Member

	mnf := &callbacks.MemberNotFound{}
	actual := mnf.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_Channel(t *testing.T) {
	var expected *discordgo.Channel

	mnf := &callbacks.MemberNotFound{}
	actual := mnf.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Is(t *testing.T) {
	cnf := &callbacks.ChannelNotFound{}

	if errors.Is(nil, &callbacks.ChannelNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.ChannelNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(cnf, &callbacks.ChannelNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestChannelNotFound_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	cnf := &callbacks.ChannelNotFound{Err: wrappedErr}

	unwrappedErr := cnf.Unwrap()

	if unwrappedErr != wrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestChannelNotFound_Error(t *testing.T) {
	cnf := &callbacks.ChannelNotFound{}
	expectedErrMsg := callbacks.ChannelNotFoundMessage

	if cnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			cnf.Error(),
			expectedErrMsg,
		)
	}

	cnf.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if cnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			cnf.Error(),
			expectedErrMsg,
		)
	}
}

func TestChannelNotFound_Guild(t *testing.T) {
	expected := &discordgo.Guild{Name: mock.TestGuild}
	cnf := &callbacks.ChannelNotFound{Guild: expected}
	actual := cnf.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Member(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	cnf := &callbacks.ChannelNotFound{Member: expected}
	actual := cnf.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Channel(t *testing.T) {
	var expected *discordgo.Channel

	cnf := &callbacks.ChannelNotFound{}
	actual := cnf.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermission_Is(t *testing.T) {
	inp := &callbacks.InsufficientPermissions{}

	if errors.Is(nil, &callbacks.InsufficientPermissions{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.InsufficientPermissions{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(inp, &callbacks.InsufficientPermissions{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestInsufficientPermission_Unwrap(t *testing.T) {
	inp := &callbacks.InsufficientPermissions{}

	unwrappedErr := inp.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestInsufficientPermission_Error(t *testing.T) {
	inp := &callbacks.InsufficientPermissions{}
	expectedErrMsg := callbacks.InsufficientPermissionMessage

	if inp.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			inp.Error(),
			expectedErrMsg,
		)
	}

	inp.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if inp.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			inp.Error(),
			expectedErrMsg,
		)
	}
}

func TestInsufficientPermissions_Guild(t *testing.T) {
	expected := &discordgo.Guild{Name: mock.TestGuild}
	inp := &callbacks.InsufficientPermissions{Guild: expected}
	actual := inp.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_Member(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	inp := &callbacks.InsufficientPermissions{Member: expected}
	actual := inp.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_Channel(t *testing.T) {
	expected := &discordgo.Channel{Name: mock.TestChannel}
	inp := &callbacks.InsufficientPermissions{Channel: expected}
	actual := inp.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func deepEqual(actual, expected interface{}) error {
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf(
			"unexpected result. Got: %+v, Expected: %+v",
			actual,
			expected,
		)
	}

	return nil
}
