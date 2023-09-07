package utils

const (
	Placed = 3
	Reject = 2
	Accept = 1
)

var OrderStatus = map[string]int{
	"placed": Placed,
	"Reject": Reject,
	"Accept": Accept,
}

// func ValidRole(val int) (string, error) {
// 	for key, value := range utils.Roles {
// 		if val == value {

// 			return key, nil
// 		}
// 	}
// 	return "", fmt.Errorf("not found")

// }
