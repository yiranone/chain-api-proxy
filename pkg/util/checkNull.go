package util

// 检查 responsePayload["result"] 是否为空或者数组大小为1
func IsResultEmptyOrSizeZeroOrEmptyObject(obj interface{}) bool {
	if obj == nil {
		return true
	}
	switch v := obj.(type) {
	case string:
		return v == "null" || v == ""
	//case []interface{}:
	//	// 如果是数组，检查是否为空或者大小为1
	//	return len(v) == 0
	case map[string]interface{}:
		// 如果是对象，检查是否为空
		return len(v) == 0
	default:
		// 其他类型可以根据需求处理，这里假设非数组和对象类型为非空
		return obj == nil
	}

	// 如果 result 不存在，视为 result 为空
	return false
}

func TruncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength]
	}
	return s
}

// 判断字符串是否在map中
func Contains(m map[string]struct{}, str string) bool {
	_, found := m[str]
	return found
}
