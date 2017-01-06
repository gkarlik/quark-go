package system

// Disposer reporesents object cleanup mechanism
type Disposer interface {
	Dispose()
}
