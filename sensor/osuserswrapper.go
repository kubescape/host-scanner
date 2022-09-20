package sensor

import (
	"io"
	"os"
	"os/user"
	"strconv"

	_ "net"
	_ "unsafe"
)

// os/users package handles extracting information from users files (/etc/passwd, /etc/group) but limited to current user root only.
// Module utilizes unexported (private) functions (using go:linkname), expanding their use for custom root path.
// NOTE: code requires environment variable CGO_ENABLED = 0

//go:linkname readColonFile os/user.readColonFile
func readColonFile(r io.Reader, fn lineFunc, readCols int) (v any, err error)

//go:linkname findUserId os/user.findUserId
func findUserId(uid string, r io.Reader) (*user.User, error)

//go:linkname findGroupId os/user.findGroupId
func findGroupId(id string, r io.Reader) (*user.Group, error)

//goLlinkname lineFunc os/user lineFunc
type lineFunc func(line []byte) (v any, err error)

const userFile = "/etc/passwd"
const groupFile = "/etc/group"

// returns *Group object if gid was found in a group file {root}/etc/group, otherwise returns nil.
func lookupGroup(gid string, root string) (*user.Group, error) {
	filePath := root + groupFile
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findGroupId(gid, f)
}

// returns group name if gid was found in a group file {root}/etc/group, otherwise returns empty string.
func LookupGroupnameByGID(gid int64, root string) (string, error) {
	groupData, err := lookupGroup(strconv.FormatInt(gid, 10), root)

	if err != nil {
		return "", err
	}

	return groupData.Name, nil

}

// returns *User object if uid was found in a users file {root}/etc/passwd, otherwise returns nil.
func lookupUser(uid string, root string) (*user.User, error) {
	filePath := root + userFile
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findUserId(uid, f)
}

// returns username if uid was found in a users file {root}/etc/passwd, otherwise returns empty string.
func LookupUsernameByUID(uid int64, root string) (string, error) {
	userData, err := lookupUser(strconv.FormatInt(uid, 10), root)

	if err != nil {
		return "", err
	}

	return userData.Username, nil

}
