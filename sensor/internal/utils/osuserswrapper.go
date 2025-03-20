package utils

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	_ "net"
	"os"
	"os/user"
	"strconv"
	"strings"
	_ "unsafe"
)

var colon = []byte{':'}

// readColonFile parses r as an /etc/group or /etc/passwd style file, running
// fn for each row. readColonFile returns a value, an error, or (nil, nil) if
// the end of the file is reached without a match.
//
// readCols is the minimum number of colon-separated fields that will be passed
// to fn; in a long line additional fields may be silently discarded.
func readColonFile(r io.Reader, fn lineFunc, readCols int) (v any, err error) {
	rd := bufio.NewReader(r)

	// Read the file line-by-line.
	for {
		var isPrefix bool
		var wholeLine []byte

		// Read the next line. We do so in chunks (as much as reader's
		// buffer is able to keep), check if we read enough columns
		// already on each step and store final result in wholeLine.
		for {
			var line []byte
			line, isPrefix, err = rd.ReadLine()

			if err != nil {
				// We should return (nil, nil) if EOF is reached
				// without a match.
				if err == io.EOF {
					err = nil
				}
				return nil, err
			}

			// Simple common case: line is short enough to fit in a
			// single reader's buffer.
			if !isPrefix && len(wholeLine) == 0 {
				wholeLine = line
				break
			}

			wholeLine = append(wholeLine, line...)

			// Check if we read the whole line (or enough columns)
			// already.
			if !isPrefix || bytes.Count(wholeLine, []byte{':'}) >= readCols {
				break
			}
		}

		// There's no spec for /etc/passwd or /etc/group, but we try to follow
		// the same rules as the glibc parser, which allows comments and blank
		// space at the beginning of a line.
		wholeLine = bytes.TrimSpace(wholeLine)
		if len(wholeLine) == 0 || wholeLine[0] == '#' {
			continue
		}
		v, err = fn(wholeLine)
		if v != nil || err != nil {
			return
		}

		// If necessary, skip the rest of the line
		for ; isPrefix; _, isPrefix, err = rd.ReadLine() {
			if err != nil {
				// We should return (nil, nil) if EOF is reached without a match.
				if err == io.EOF {
					err = nil
				}
				return nil, err
			}
		}
	}
}
func matchGroupIndexValue(value string, idx int) lineFunc {
	var leadColon string
	if idx > 0 {
		leadColon = ":"
	}
	substr := []byte(leadColon + value + ":")
	return func(line []byte) (v any, err error) {
		if !bytes.Contains(line, substr) || bytes.Count(line, colon) < 3 {
			return
		}
		// wheel:*:0:root
		parts := strings.SplitN(string(line), ":", 4)
		if len(parts) < 4 || parts[0] == "" || parts[idx] != value ||
			// If the file contains +foo and you search for "foo", glibc
			// returns an "invalid argument" error. Similarly, if you search
			// for a gid for a row where the group name starts with "+" or "-",
			// glibc fails to find the record.
			parts[0][0] == '+' || parts[0][0] == '-' {
			return
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			return nil, nil
		}
		return &user.Group{Name: parts[0], Gid: parts[2]}, nil
	}
}

// returns a *User for a row if that row's has the given value at the
// given index.
func matchUserIndexValue(value string, idx int) lineFunc {
	var leadColon string
	if idx > 0 {
		leadColon = ":"
	}
	substr := []byte(leadColon + value + ":")
	return func(line []byte) (v any, err error) {
		if !bytes.Contains(line, substr) || bytes.Count(line, colon) < 6 {
			return
		}
		// kevin:x:1005:1006::/home/kevin:/usr/bin/zsh
		parts := strings.SplitN(string(line), ":", 7)
		if len(parts) < 6 || parts[idx] != value || parts[0] == "" ||
			parts[0][0] == '+' || parts[0][0] == '-' {
			return
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			return nil, nil
		}
		if _, err := strconv.Atoi(parts[3]); err != nil {
			return nil, nil
		}
		u := &user.User{
			Username: parts[0],
			Uid:      parts[2],
			Gid:      parts[3],
			Name:     parts[4],
			HomeDir:  parts[5],
		}
		// The pw_gecos field isn't quite standardized. Some docs
		// say: "It is expected to be a comma separated list of
		// personal data where the first item is the full name of the
		// user."
		u.Name, _, _ = strings.Cut(u.Name, ",")
		return u, nil
	}
}

func findUserId(uid string, r io.Reader) (*user.User, error) {
	i, e := strconv.Atoi(uid)
	if e != nil {
		return nil, errors.New("user: invalid userid " + uid)
	}
	if v, err := readColonFile(r, matchUserIndexValue(uid, 2), 6); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*user.User), nil
	}
	return nil, user.UnknownUserIdError(i)
}

func findGroupId(id string, r io.Reader) (*user.Group, error) {
	if v, err := readColonFile(r, matchGroupIndexValue(id, 2), 3); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*user.Group), nil
	}
	return nil, user.UnknownGroupIdError(id)
}

type lineFunc func(line []byte) (v any, err error)

const userFile = "/etc/passwd"
const groupFile = "/etc/group"

var (
	userGroupCache = map[string]userGroupCacheItem{} // map[rootDir]struct{users, groups}
)

type userGroupCacheItem struct {
	users  map[string]string
	groups map[string]string
}

// getUserName checks if uid is cached, if not, it tries to find it in a users file {root}/etc/passwd.
func getUserName(uid int64, root string) (string, error) {

	// return from cache if exists
	if users, ok := userGroupCache[root]; ok {
		if username, ok := users.users[strconv.Itoa(int(uid))]; ok {
			return username, nil
		}
	}

	// find username in a users file
	username, err := lookupUsernameByUID(uid, root)
	if err != nil {
		return "", err
	}

	// cache username
	if _, ok := userGroupCache[root]; !ok {
		userGroupCache[root] = userGroupCacheItem{
			users:  map[string]string{},
			groups: map[string]string{},
		}
	}

	userGroupCache[root].users[strconv.Itoa(int(uid))] = username

	return username, nil
}

// getGroupName checks if gid is cached, if not, it tries to find it in a group file {root}/etc/group.
func getGroupName(gid int64, root string) (string, error) {

	// return from cache if exists
	if users, ok := userGroupCache[root]; ok {
		if groupname, ok := users.groups[strconv.Itoa(int(gid))]; ok {
			return groupname, nil
		}
	}

	// find groupname in a group file
	groupname, err := LookupGroupnameByGID(gid, root)
	if err != nil {
		return "", err
	}

	// cache groupname
	if _, ok := userGroupCache[root]; !ok {
		userGroupCache[root] = userGroupCacheItem{
			users:  map[string]string{},
			groups: map[string]string{},
		}
	}

	userGroupCache[root].groups[strconv.Itoa(int(gid))] = groupname

	return groupname, nil
}

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
func lookupUsernameByUID(uid int64, root string) (string, error) {
	userData, err := lookupUser(strconv.FormatInt(uid, 10), root)

	if err != nil {
		return "", err
	}

	return userData.Username, nil
}
