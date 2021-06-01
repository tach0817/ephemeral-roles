// Package operations provides a centralized gateway for processing requests
// on Discord API operations.
package operations

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/sync/singleflight"
)

// APIErrorCodeMaxRoles is the Discord API error code for max roles.
const APIErrorCodeMaxRoles = 30005

const guildMembersPageLimit = 1000

// LookupGuild returns a *discordgo.Guild from the session's internal state
// cache. If the guild is not found in the state cache, LookupGuild will query
// the Discord API for the guild and add it to the state cache before returning
// it.
func LookupGuild(flightGroup *singleflight.Group, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	object, err, _ := flightGroup.Do(guildID, func() (interface{}, error) {
		guild, err := session.State.Guild(guildID)
		if err != nil {
			return updateStateGuilds(session, guildID)
		}

		return guild, nil
	})
	if err != nil {
		return nil, err
	}

	guild, ok := object.(*discordgo.Guild)
	if !ok {
		return nil, fmt.Errorf(
			"unable to type assert to *discordgo.Guild: %T",
			object,
		)
	}

	return guild, nil
}

// AddRoleToMember adds the role associated with the provided roleID to the
// user associated with the provided userID, in the guild associated with the
// provided guildID.
func AddRoleToMember(session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleAdd(guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to add ephemeral role: %w", err)
	}

	return nil
}

// RemoveRoleFromMember removes the role associated with the provided roleID
// from the user associated with the provided userID, in the guild associated
// with the provided guildID.
func RemoveRoleFromMember(session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleRemove(guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to remove ephemeral role: %w", err)
	}

	return nil
}

// IsDeadlineExceeded checks if the provided error wraps
// context.DeadlineExceeded.
func IsDeadlineExceeded(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

// IsForbiddenResponse checks if the provided error wraps *discordgo.RESTError.
// If it does, IsForbiddenResponse returns true if the response code is equal
// to http.StatusForbidden.
func IsForbiddenResponse(err error) bool {
	var restErr *discordgo.RESTError

	if errors.As(err, &restErr) {
		if restErr.Response.StatusCode == http.StatusForbidden {
			return true
		}
	}

	return false
}

// IsMaxGuildsResponse checks if the provided error wraps *discordgo.RESTError.
// If it does, IsMaxGuildsResponse returns true if the response code is equal
// to http.StatusBadRequest and the error code is 30005.
func IsMaxGuildsResponse(err error) bool {
	var restErr *discordgo.RESTError

	if errors.As(err, &restErr) {
		if restErr.Response.StatusCode == http.StatusBadRequest {
			return restErr.Message.Code == APIErrorCodeMaxRoles
		}
	}

	return false
}

// ShouldLogDebug checks if the provided error should be logged at a debug
// level.
func ShouldLogDebug(err error) bool {
	switch {
	case IsDeadlineExceeded(err), IsForbiddenResponse(err):
		return true
	default:
		return false
	}
}

// BotHasChannelPermission checks if the bot has view permissions for the
// channel. If the bot does have the view permission, BotHasChannelPermission
// returns nil.
func BotHasChannelPermission(session *discordgo.Session, channel *discordgo.Channel) error {
	permissions, err := session.UserChannelPermissions(session.State.User.ID, channel.ID)
	if err != nil {
		return fmt.Errorf("unable to determine channel permissions: %w", err)
	}

	if permissions&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
		return fmt.Errorf("insufficient channel permissions: channel: %s", channel.Name)
	}

	return nil
}

func updateStateGuilds(session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("error sending guild query request: %w", err)
	}

	roles, err := session.GuildRoles(guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	channels, err := session.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	members, err := recursiveGuildMembers(session, guildID, "", guildMembersPageLimit)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild members: %w", err)
	}

	guild.Roles = roles
	guild.Channels = channels
	guild.Members = members
	guild.MemberCount = len(members)

	err = session.State.GuildAdd(guild)
	if err != nil {
		return nil, fmt.Errorf("unable to add guild to state cache: %w", err)
	}

	return guild, nil
}

func recursiveGuildMembers(
	session *discordgo.Session,
	guildID, after string,
	limit int,
) ([]*discordgo.Member, error) {
	guildMembers, err := session.GuildMembers(guildID, after, limit)
	if err != nil {
		return nil, fmt.Errorf("error sending recursive guild members request: %w", err)
	}

	if len(guildMembers) < guildMembersPageLimit {
		return guildMembers, nil
	}

	nextGuildMembers, err := recursiveGuildMembers(
		session,
		guildID,
		guildMembers[len(guildMembers)-1].User.ID,
		guildMembersPageLimit,
	)
	if err != nil {
		return nil, err
	}

	return append(guildMembers, nextGuildMembers...), nil
}
