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
