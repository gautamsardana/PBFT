package logic

const (
	StInit        = 100
	StPrePrepared = 101
	StPrepared    = 102
	StCommitted   = 103
	StExecuted    = 104
	StFailed      = 105
	StNoOp        = 110
)

var StateToStringMap = map[int32]string{
	StInit:        "Init",
	StPrePrepared: "Pre-prepared",
	StPrepared:    "Prepared",
	StCommitted:   "Committed",
	StExecuted:    "Executed",
	StFailed:      "Failed",
	StNoOp:        "No-op",
}
