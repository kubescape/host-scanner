package sensor

type ActionType int

const (
	ActionTypeGetKubeletCMD = iota + 1
)

var (
	// Where the host sensor is expecting host fs to be mounted.
	// Defined as var for testing purposes only
	hostFileSystemDefaultLocation = "/host_fs"
)
