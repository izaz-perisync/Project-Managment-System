package utils

const (
	Dispatch  = 5
	Delivered = 4
	Placed    = 3
	Reject    = 2
	Accept    = 1
)

var Active = map[string]int{
	"placed":   Placed,
	"dispatch": Dispatch,
	"accept":   Accept,
	"reject":   Reject,
	"deliver":  Delivered,
}

var OrderStatus = map[string]any{
	"Active": "Active",
	"Closed": "Closed",
}

