package vo

// EventType is a broker routing key for avatar lifecycle events.
type EventType string

const (
	EventAvatarUploaded EventType = "avatar.uploaded"
	EventAvatarDeleted  EventType = "avatar.deleted"
)
