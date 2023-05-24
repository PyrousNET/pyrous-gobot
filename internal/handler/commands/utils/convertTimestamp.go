package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertTimestamp(in string) (int64, error) {
	parts := strings.Split(in, "-")
	if len(parts) < 2 {
		return 0, fmt.Errorf("parts of payload were inccorect for pubsub")
	}
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)

	return timestamp, err
}
