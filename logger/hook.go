// Copyright 2020 Pradyumna Kaushik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
)

// WriterHook is a hook that writes logs, that contain at least one of the specified set of topics
// as keys in its fields, to the specified Writer. The logs are formatted using the specified formatter.
type WriterHook struct {
	formatter logrus.Formatter
	writer    io.Writer
	topics    map[string]struct{}
}

// newWriterHook instantiates and returns a new WriterHook.
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

// Levels return the list of levels for which this hook will be fired.
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

// Fire checks whether the fields in the provided entry contain at least one of the specified
// topics and if yes, formats the entry using the specified formatter and then writes it to the
// specified Writer.
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
