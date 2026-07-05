package service

import "ostenia/internal/backend/interfaces"

type MockSystem struct {
	PathUser    string
	PathMachine string
	PIDs        map[string][]int
	Ports       map[int][]int
	IsAdminVal  bool
}

func NewMockSystem() *MockSystem {
	return &MockSystem{
		PIDs:  make(map[string][]int),
		Ports: make(map[int][]int),
	}
}

func (m *MockSystem) GetPath(target string) (string, error) {
	if target == "Machine" {
		return m.PathMachine, nil
	}
	return m.PathUser, nil
}

func (m *MockSystem) SetPath(path string, target string) error {
	if target == "Machine" {
		m.PathMachine = path
	} else {
		m.PathUser = path
	}
	return nil
}

func (m *MockSystem) FindOsteniaPIDs(exeName string) []int {
	return m.PIDs[exeName]
}

func (m *MockSystem) FindPortsByPID(pid int) []int {
	return m.Ports[pid]
}

func (m *MockSystem) KillProcess(pid int) error {
	return nil
}

func (m *MockSystem) CreateSymlink(target, link string) error {
	return nil
}

func (m *MockSystem) IsAdmin() bool {
	return m.IsAdminVal
}

func (m *MockSystem) NotifyEnvironmentUpdate() {}

var _ interfaces.System = (*MockSystem)(nil)
