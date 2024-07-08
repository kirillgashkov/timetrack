package timetrackdb

// TrackTypeStart and TrackTypeStop are the types of time tracking events.
const (
	TrackTypeStart = "start"
	TrackTypeStop  = "stop"
)

// TaskUserStatusActive and TaskUserStatusInactive are the statuses of a user's
// current participation in a task.
const (
	TaskUserStatusActive   = "active"
	TaskUserStatusInactive = "inactive"
)
