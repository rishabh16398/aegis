package network

// Monitor will be implemented in Task 7.
type Monitor struct{}

func New(iface string) *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start() error { return nil }
func (m *Monitor) Stop()        {}
