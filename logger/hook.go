package logger

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
)

type WriterHook struct {
	formatter logrus.Formatter
	writer    io.Writer
	topics    map[string]struct{}
}

func newWriterHook(formatter logrus.Formatter, writer io.Writer, topics ...string) logrus.Hook {
	hook := &WriterHook{
		formatter: formatter,
		writer:    writer,
		topics:    make(map[string]struct{}),
	}
	for _, topic := range topics {
		hook.topics[topic] = struct{}{}
	}

	return hook
}

func (h WriterHook) Levels() []logrus.Level {
	// We do not want debug and trace level logs to be persisted as they are typically temporary.
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

func (h *WriterHook) Fire(entry *logrus.Entry) error {
	// Logging only if any of the provided topics are found as keys in fields.
	allow := false
	for topic := range h.topics {
		if _, ok := entry.Data[topic]; ok {
			allow = true
			break
		}
	}

	var err error
	var formattedLog []byte
	if allow {
		formattedLog, err = h.formatter.Format(entry)
		if err != nil {
			err = errors.Wrap(err, "failed to format entry")
		} else {
			_, err = h.writer.Write(formattedLog)
		}
	}
	return err
}
