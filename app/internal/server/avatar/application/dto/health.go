package dto

// HealthCheckOutput is application output with component statuses.
type HealthCheckOutput struct {
	Status   string
	Database string
	Storage  string
	Broker   string
}
