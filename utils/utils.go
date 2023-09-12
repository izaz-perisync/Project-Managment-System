package utils

const (
	Dispatch  = 5
	Delivered = 4
	Placed    = 3
	Reject    = 2
	Accept    = 1
)

var Active = map[string]string{
	"placed":   "placed",
	"dispatch": "dispatch",
	"accept":   "accept",
	"reject":   "reject",
	"deliver":  "deliver",
}

var OrderStatus = map[string]any{
	"Active": "Active",
	"Closed": "Closed",
}
