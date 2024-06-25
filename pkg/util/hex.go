package util

import (
	"fmt"
	"strconv"
)

func IncrementHex(hexStr string) (string, error) {
	// 将十六进制字符串转换为整数
	value, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return "", err
	}

	// 对整数加1
	value++

	// 将结果转换回十六进制字符串
	// 使用 fmt.Sprintf 来格式化为十六进制字符串，并去掉前缀 "0x"
	hexResult := fmt.Sprintf("0x%x", value)

	return hexResult, nil
}
