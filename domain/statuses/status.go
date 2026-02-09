package statuses

type IdempotencyStatus string

const (
	Processing IdempotencyStatus = "processing"
	Error      IdempotencyStatus = "error"
	Done       IdempotencyStatus = "done"
)
