/**
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package strategies

type Strategy interface {
	// SetTaskRanksReceiver registers a receiver of the task ranking results.
	// This receiver is a callback and is used to pass the result of applying
	// the strategy to rank tasks.
	SetTaskRanksReceiver(TaskRanksReceiver)
	// Execute the strategy.
	Execute()
}

// Build the strategy object.
func Build(s Strategy, receiver TaskRanksReceiver) {
	s.SetTaskRanksReceiver(receiver)
}
