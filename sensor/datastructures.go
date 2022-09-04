package sensor

type LinuxSecurityHardeningStatus struct {
	AppArmor string `json:"appArmor"`
	SeLinux  string `json:"seLinux"`
}

// FileInfo holds information about a file
type FileInfo struct {
	// Ownership information
	Ownership *FileOwnership `json:"ownership"`

	// The path of the file
	// Example: /etc/kubernetes/manifests/kube-apiserver.yaml
	Path string `json:"path"`

	// Content of the file
	Content     []byte `json:"content,omitempty"`
	Permissions int    `json:"permissions"`
}

// FileOwnership holds the ownership of a file
type FileOwnership struct {
	Err string `json:"err,omitempty"`
	UID int64  `json:"uid"`
	GID int64  `json:"gid"`
}
