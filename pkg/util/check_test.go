package util

import "testing"

type GenericTestJSON map[string]interface{}

func TestCurrentBoardInfo(t *testing.T) {
	var responsePayload GenericTestJSON = make(map[string]interface{})
	responsePayload["result"] = []interface{}{}
	result := IsResultEmptyOrSizeZeroOrEmptyObject(responsePayload["result"])
	t.Logf("Manufacturer:%v", result)

	result2 := IsResultEmptyOrSizeZeroOrEmptyObject("null")
	t.Logf("Manufacturer2:%v", result2)
}
