package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"sort"
	"strings"
)

type MyLogFormatter struct{}

func (m *MyLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")

	//var RequestId string
	//if entry.Context != nil {
	//	RequestId = fmt.Sprintf("%s", entry.Context.Value("tid"))
	//}

	var dataStr string
	if len(entry.Data) > 0 {
		data := make(logrus.Fields)
		for k, v := range entry.Data {
			data[k] = v
		}
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var sb *bytes.Buffer
		sb = &bytes.Buffer{}
		for _, key := range keys {
			var value interface{}
			switch {
			default:
				value = data[key]
			}
			m.appendKeyValue(sb, key, value)
		}
		if sb != nil {
			dataStr = string(sb.Bytes())
		}
	}

	var newLog string
	if entry.HasCaller() {
		fName := filepath.Base(entry.Caller.File)
		lineNum := entry.Caller.Line
		functionMsg := entry.Caller.Function
		functionMsg = strings.ReplaceAll(functionMsg, "github.com/yiranone/", "")
		newLog = fmt.Sprintf("[%s] [%5s] [%s:%d %s] [%s] %s\n",
			timestamp, entry.Level, fName, lineNum, functionMsg, dataStr, entry.Message)
	} else {
		newLog = fmt.Sprintf("[%s] [%5s] [%s] %s\n",
			timestamp, entry.Level, dataStr, entry.Message)
	}
	b.WriteString(newLog)
	return b.Bytes(), nil
}

func (m *MyLogFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	m.appendValue(b, value)
}

func (m *MyLogFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	//if !m.needsQuoting(stringVal) {
	b.WriteString(stringVal)
	//} else {
	//	b.WriteString(fmt.Sprintf("%q", stringVal))
	//}
}
