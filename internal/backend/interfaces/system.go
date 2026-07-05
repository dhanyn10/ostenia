package interfaces

// System abstracts OS-level operations for cross-platform compatibility and testing.
type System interface {
	GetPath(target string) (string, error)
	SetPath(path string, target string) error
	FindOsteniaPIDs(exeName string) []int
	FindPortsByPID(pid int) []int
	KillProcess(pid int) error
	CreateSymlink(target, link string) error
	IsAdmin() bool
	NotifyEnvironmentUpdate()
}
