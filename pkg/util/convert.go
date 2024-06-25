package util

import (
	"fmt"
	"strconv"
)

func ConvertToInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("value not found")
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type for value")
	}
}

func ConvertToString(value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("value not found")
	}

	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(v), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported type for value")
	}
}
