package logger

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// CSVFormatter is a log formatter that formats the logs as a CSV.
// The keys in the fields are used as headers and the field values in subsequent
// calls to Format() will be matched and ordered based on the columns in the csv.
type CSVFormatter struct {
	 headers []string
}

func NewCSVFormatter() logrus.Formatter {
	return &CSVFormatter{headers: make([]string, 0)}
}

func (f *CSVFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buf = &bytes.Buffer{}
	if len(f.headers) == 0 {
		for h, _ := range entry.Data {
			f.headers = append(f.headers, h)
		}
		buf.WriteString(strings.Join(f.headers, ","))
		buf.WriteByte('\n')
	}

	// values.
	var values []string
	for _, h := range f.headers {
		fmt.Println(h)
		fmt.Println(entry.Data)
		values = append(values, entry.Data[h].(string))
	}

	buf.WriteString(strings.Join(values, ","))
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
