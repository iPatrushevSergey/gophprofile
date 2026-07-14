package vo

// OutboxEventType describes an outbox event kind.
type OutboxEventType string

const (
	OutboxEventAvatarUploaded OutboxEventType = "avatar.uploaded"
	OutboxEventAvatarDeleted  OutboxEventType = "avatar.deleted"
)
