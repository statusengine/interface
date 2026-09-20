// Package domain holds the types that cross the boundary between the
// repositories and the HTTP handlers.
package domain

// HostState is Naemon's host state. The numbers are the core's own, not
// ours, so they are safe to compare against raw table values.
type HostState int

const (
	HostUp          HostState = 0
	HostDown        HostState = 1
	HostUnreachable HostState = 2
)

// ServiceState is Naemon's service state.
type ServiceState int

const (
	ServiceOK       ServiceState = 0
	ServiceWarning  ServiceState = 1
	ServiceCritical ServiceState = 2
	ServiceUnknown  ServiceState = 3
)

// StateText names a host state for a client that should not have to keep
// its own copy of the mapping.
func (s HostState) String() string {
	switch s {
	case HostUp:
		return "up"
	case HostDown:
		return "down"
	case HostUnreachable:
		return "unreachable"
	default:
		return "unknown"
	}
}

func (s ServiceState) String() string {
	switch s {
	case ServiceOK:
		return "ok"
	case ServiceWarning:
		return "warning"
	case ServiceCritical:
		return "critical"
	case ServiceUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// Kind distinguishes the two things that can have a problem, for the
// lists that mix them.
type Kind string

const (
	KindHost    Kind = "host"
	KindService Kind = "service"
)
