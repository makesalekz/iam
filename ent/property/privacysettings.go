package property

type PrivacySettings string

const (
	MyLastActions   PrivacySettings = "MY_LAST_ACTIONS"
	MyProfileImage  PrivacySettings = "MY_PROFILE_IMAGE"
	MyEvents        PrivacySettings = "MY_EVENTS"
	GroupChatInvite PrivacySettings = "GROUP_CHAT_INVITE"
	EventInvite     PrivacySettings = "EVENT_INVITE"
)

// Values provides list valid values for Enum.
func (PrivacySettings) Values() (kinds []string) {
	for _, s := range []PrivacySettings{MyLastActions, MyProfileImage, MyEvents, GroupChatInvite, EventInvite} {
		kinds = append(kinds, string(s))
	}
	return
}
