package enum

type PrivacySettings string

const (
	MyLastActions   PrivacySettings = "MY_LAST_ACTIONS"
	MyProfileImage  PrivacySettings = "MY_PROFILE_IMAGE"
	MyEvents        PrivacySettings = "MY_EVENTS"
	GroupChatInvite PrivacySettings = "GROUP_CHAT_INVITE"
	EventInvite     PrivacySettings = "EVENT_INVITE"
	MySlots         PrivacySettings = "MY_SLOTS"
	SlotsDetails    PrivacySettings = "SLOTS_DETAILS"
	LastVisit       PrivacySettings = "LAST_VISIT"
)

// Values provides list valid values for Enum.
func (PrivacySettings) Values() (kinds []string) {
	for _, s := range []PrivacySettings{MyLastActions, MyProfileImage, MyEvents, GroupChatInvite, EventInvite, MySlots, SlotsDetails, LastVisit} {
		kinds = append(kinds, string(s))
	}
	return
}
