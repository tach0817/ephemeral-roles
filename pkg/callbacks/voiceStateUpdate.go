package callbacks

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultRoleColor = 16753920 // Default to orange hex #FFA500 in decimal

type vsuEvent struct {
	Session     *discordgo.Session
	User        *discordgo.User
	Guild       *discordgo.Guild
	GuildRoles  discordgo.Roles
	MemberRoles discordgo.Roles
}

// VoiceStateUpdate is the callback function for the VoiceStateUpdate event from Discord
func (config *Config) VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	// Increment the total number of VoiceStateUpdate events
	config.VoiceStateUpdateCounter.Inc()

	event := &vsuEvent{
		Session: s,
	}
	// Get the user
	user, err := s.User(vsu.UserID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to determine user in VoiceStateUpdate")

		return
	}

	event.User = user

	// Get the guild
	guild, err := s.Guild(vsu.GuildID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to determine guild in VoiceStateUpdate")

		return
	}

	event.Guild = guild

	// Get the guild's roles
	guildRoles, dErr := guildRoles(s, vsu.GuildID)
	if dErr != nil {
		config.Log.WithError(dErr).Debugf("Unable to determine guild roles in VoiceStateUpdate")

		return
	}

	event.GuildRoles = guildRoles

	// Get the guild member's roles
	memberRoles, err := guildMemberRoles(event)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("Unable to determine guild member roles")

		return
	}

	event.MemberRoles = memberRoles

	// Check if user disconnect event
	if vsu.ChannelID == "" {
		config.revokeEphemeralRoles(event)

		config.Log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		return
	}

	// Get the channel
	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("Unable to determine channel in VoiceStateUpdate")

		return
	}

	ephRoleName := config.RolePrefix + " " + channel.Name

	// Check to see if the role already exists in the guild
	for _, guildRole := range guildRoles {
		if guildRole.Name != ephRoleName {
			continue
		}

		// Check to see if the member already has the role
		for _, memberRole := range memberRoles {
			if memberRole.ID == guildRole.ID {
				return // No effective change
			}
		}

		// Ephemeral role exists, add member to it
		config.grantEphemeralRole(event, guildRole)

		return
	}

	// Ephemeral role does not exist, create and edit it
	ephRole, err := config.guildRoleCreateEdit(event, ephRoleName)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"role":  ephRoleName,
			"guild": guild.Name,
		}).Debugf("Unable to manage ephemeral role")

		return
	}

	// Add role to member
	config.grantEphemeralRole(event, ephRole)
}

// guildRoles handles role lookups using dErr *discordError as a means to
// provide context to API errors
func guildRoles(s *discordgo.Session, guildID string) (roles []*discordgo.Role, dErr *discordError) {
	var err error

	roles, err = s.GuildRoles(guildID)
	if err != nil {
		// Find the JSON with regular expressions
		rx := regexp.MustCompile("{.*}")
		errHTTPString := rx.ReplaceAllString(err.Error(), "")
		errJSONString := rx.FindString(err.Error())

		dAPIResp := &DiscordAPIResponse{}

		dErr = &discordError{
			HTTPResponseMessage: errHTTPString,
			APIResponse:         dAPIResp,
			CustomMessage:       "",
		}

		unmarshalErr := json.Unmarshal([]byte(errJSONString), dAPIResp)
		if unmarshalErr != nil {
			dAPIResp.Code = -1
			dAPIResp.Message = "Unable to unmarshal Discord API JSON response: " + errJSONString
			return
		}

		// Add CustomMessage as appropriate
		switch dErr.APIResponse.Code {
		case 50013: // Code 50013: "Missing Permissions"
			dErr.CustomMessage = "Insufficient role permission to query guild roles"
		}
	}

	return
}

func guildMemberRoles(event *vsuEvent) ([]*discordgo.Role, error) {
	// Get guild member
	guildMember, err := event.Session.GuildMember(event.Guild.ID, event.User.ID)
	if err != nil {
		return make([]*discordgo.Role, 0),
			errors.Wrap(err, "unable to determine member in VoiceStateUpdate: "+err.Error())
	}

	// Map our member roles
	memberRoleIDs := make(map[string]bool)
	for _, roleID := range guildMember.Roles {
		memberRoleIDs[roleID] = true
	}

	memberRoles := make([]*discordgo.Role, 0)

	for _, role := range event.GuildRoles {
		if memberRoleIDs[role.ID] {
			memberRoles = append(memberRoles, role)
		}
	}

	return memberRoles, nil
}

func (config *Config) guildRoleCreateEdit(event *vsuEvent, ephRoleName string) (*discordgo.Role, error) {
	// Create a new blank role
	ephRole, err := event.Session.GuildRoleCreate(event.Guild.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ephemeral role: "+err.Error())
	}

	roleColor := defaultRoleColor

	// Check for role color override
	if colorString, found := os.LookupEnv("ROLE_COLOR_HEX2DEC"); found {
		roleColor, err = strconv.Atoi(colorString)
		if err != nil {
			config.Log.
				WithError(err).
				WithField("ROLE_COLOR_HEX2DEC", colorString).
				Warnf("Error parsing ROLE_COLOR_HEX2DEC from environment")

			roleColor = defaultRoleColor
		}
	}

	// Edit the new role
	ephRole, err = event.Session.GuildRoleEdit(
		event.Guild.ID,
		ephRole.ID,
		ephRoleName,
		roleColor,
		true,
		ephRole.Permissions,
		ephRole.Mentionable,
	)
	if err != nil {
		return nil, errors.New("unable to edit ephemeral role: " + err.Error())
	}

	/*err = guildRolesReorder(s, guild.ID)
	if err != nil {
		return nil, errors.New("unable to reorder ephemeral role: " + err.Error())
	}*/

	return ephRole, nil
}

func (config *Config) revokeEphemeralRoles(event *vsuEvent) {
	for _, role := range event.MemberRoles {
		if strings.HasPrefix(role.Name, config.RolePrefix) {
			// Found ephemeral role, revoke it
			err := event.Session.GuildMemberRoleRemove(event.Guild.ID, event.User.ID, role.ID)
			if err != nil {
				config.Log.WithError(err).
					WithFields(logrus.Fields{
						"user":  event.User.Username,
						"guild": event.Guild.Name,
						"role":  role.Name,
					}).Debugf("Unable to remove role on VoiceStateUpdate")

				return
			}

			config.Log.WithFields(logrus.Fields{
				"user":  event.User.Username,
				"guild": event.Guild.Name,
				"role":  role.Name,
			}).Debugf("Removed role")
		}
	}
}

func (config *Config) grantEphemeralRole(event *vsuEvent, ephRole *discordgo.Role) {
	// Revoke any previous ephemeral roles
	config.revokeEphemeralRoles(event)

	// Add our member to role
	err := event.Session.GuildMemberRoleAdd(event.Guild.ID, event.User.ID, ephRole.ID)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"user":  event.User.Username,
			"role":  ephRole.Name,
			"guild": event.Guild.Name,
		}).Debugf("Unable to add user to ephemeral role")

		return
	}

	config.Log.WithFields(logrus.Fields{
		"user":  event.User.Username,
		"role":  ephRole.Name,
		"guild": event.Guild.Name,
	}).Debugf("Added role")
}

/*func guildRolesReorder(s *discordgo.Session, guildID string) error {
	guildRoles, dErr := guildRoles(s, guildID)
	if dErr != nil {
		return errors.New(dErr.Error())
	}

	roles := orderedRoles(guildRoles)

	log.WithField("roles", roles).Debugf("Old role order")

	sort.SliceStable(
		roles,
		func(i, j int) bool {
			return roles[i].Position < roles[j].Position
		},
	)

	// Alignment correction if Discord is slow to update
	for index, role := range roles {
		if role.Position != index {
			role.Position = index
		}
	}

	for index, role := range roles {
		if role.Name == "@everyone" && role.Position != 0 { // @everyone should be the lowest
			roles.swap(index, 0)
		}

		if role.Name == BOTNAME && role.Position != len(roles)-1 { // BOTNAME should be the highest
			roles.swap(index, len(roles)-1)
		}
	}

	// Bubble the ephemeral roles up
	for index, role := range roles {
		if strings.HasPrefix(role.Name, ROLEPREFIX) {
			for j := index; j < len(roles)-2; j++ {
				// Stop bubbling at the bottom of the top-most group
				if !strings.HasPrefix(roles[j+1].Name, ROLEPREFIX) {
					roles.swap(j, j+1)
				}
			}
		}
	}

	log.WithField("roles", roles).Debugf("New role order")

	_, err := s.GuildRoleReorder(guildID, roles)
	if err != nil {
		err = errors.New("unable to reorder guild roles from API: " + err.Error())

		return err
	}

	return nil
}*/
