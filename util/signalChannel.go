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

package util

import "sync"

// SignalChannel represents a wrapper on channels used for sending signals (type chan struct{}).
// This wrapper abstracts away the select blocks used, makes the operations
// safe (such as ensuring that the channels are closed only once) and exposes simple functions
// that can be used for common operations performed on channels.
// Cite: https://go101.org/article/channel-closing.html
type SignalChannel struct {
	C chan struct{}
	// Storing pointer of sync.Once to ensure that lock is not passed by value (sync.Once holds sync.Mutex).
	once *sync.Once
}

// NewSignalChannel returns a new channel that can be used for signaling.
func NewSignalChannel() *SignalChannel {
	return &SignalChannel{
		C:    make(chan struct{}),
		once: &sync.Once{},
	}
}

// Close the signal channel. As closing a closed channel panics, this function ensures
// that the channel is closed only once.
func (s *SignalChannel) Close() {
	// Making sure that the channel is closed only once.
	s.once.Do(func() {
		close(s.C)
	})
}

// IsClosed performs a non-blocking check on whether the channel is closed.
func (s *SignalChannel) IsClosed() bool {
	select {
	case <-s.C:
		return true
	default:
		return false
	}
}

// WaitTillClosed blocks until the channel is closed and then returns.
func (s *SignalChannel) WaitTillClosed() {
	<-s.C
}
