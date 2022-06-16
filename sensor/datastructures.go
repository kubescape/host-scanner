package sensor

type LinuxSecurityHardeningStatus struct {
	AppArmor string `json:"appArmor"`
	SeLinux  string `json:"seLinux"`
}

// FileInfo holds information about a file
type FileInfo struct {
	// The path of the file
	// Example: /etc/kubernetes/manifests/kube-apiserver.yaml
	Path string `json:"path"`

	// Ownership information
	Ownership *FileOwnership `json:"ownership"`

	// File permissions as integer (UNIX permissions)
	// Example: 438
	Permissions int `json:"permissions"`

	// Content of the file
	Content []byte `json:"content"`
}

// FileOwnership holds the ownership of a file
type FileOwnership struct {
	// The user who owns the file
	// Example: root
	UID int64 `json:"uid"`

	// The group to which the file belongs
	// Example: root
	GID int64 `json:"gid"`

	// The error that prevent fetching of the ownership of the file (if any)
	// Example: file not exist
	Err string `json:"err,omitempty"`
}
