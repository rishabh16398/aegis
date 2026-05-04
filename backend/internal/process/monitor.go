package process

// Monitor will be implemented in Task 8.
type Monitor struct{}

func New() *Monitor { return &Monitor{} }

func (m *Monitor) Start() {}
func (m *Monitor) Stop()  {}
