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

package strategies

import "log"

// TaskRankCpuSharesStrategy is a task ranking strategy that ranks the tasks
// in non-increasing order based on the cpu-shares allocated to tasks.
type TaskRankCpuSharesStrategy struct {
	receiver TaskRanksReceiver
}

// SetTaskRanksReceiver sets the receiver of the results of task ranking.
func (s *TaskRankCpuSharesStrategy) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
	s.receiver = receiver
}

func (s *TaskRankCpuSharesStrategy) Execute(data string) {
	// placeholder.
	s.receiver.Receive(nil)
	log.Println("from::cpushares_strategy " + data)
}
