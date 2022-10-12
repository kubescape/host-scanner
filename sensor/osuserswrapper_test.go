package sensor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserName(t *testing.T) {
	userGroupCache = map[string]userGroupCacheItem{}
	// regular
	t.Run("regular", func(t *testing.T) {
		name, _ := getUserName(0, "testdata")
		assert.Equal(t, "root", name)
		assert.Contains(t, userGroupCache, "testdata")
		assert.Contains(t, userGroupCache["testdata"].users, "0")
		assert.Equal(t, "root", userGroupCache["testdata"].users["0"])
	})

	// cached
	t.Run("cached", func(t *testing.T) {
		userGroupCache["foo"] = userGroupCacheItem{
			users:  map[string]string{"0": "bar"},
			groups: map[string]string{},
		}
		name, _ := getUserName(0, "foo")
		assert.Equal(t, "bar", name)
	})
}

func TestGetGroupName(t *testing.T) {
	userGroupCache = map[string]userGroupCacheItem{}

	// regular
	t.Run("regular", func(t *testing.T) {
		name, _ := getGroupName(0, "testdata")
		assert.Equal(t, "root", name)
		assert.Contains(t, userGroupCache, "testdata")
		assert.Contains(t, userGroupCache["testdata"].groups, "0")
		assert.Equal(t, "root", userGroupCache["testdata"].groups["0"])
	})

	// cached
	t.Run("cached", func(t *testing.T) {
		userGroupCache["foo"] = userGroupCacheItem{
			users:  map[string]string{},
			groups: map[string]string{"0": "bar"},
		}
		name, _ := getGroupName(0, "foo")
		assert.Equal(t, "bar", name)
	})
}

func Test_LookupUsernameByUID(t *testing.T) {
	uid_tests := []struct {
		name        string
		root        string
		uid         int64
		expectedRes string
		wantErr     bool
	}{
		{
			name:        "testdata_uid_exists",
			root:        "testdata",
			uid:         0,
			expectedRes: "root",
			wantErr:     false,
		},
		{
			name:        "testdata_uid_not_exists",
			root:        "testdata",
			uid:         10,
			expectedRes: "root",
			wantErr:     true,
		},
		{
			name:        "testdata_file_not_exists",
			root:        "testdata/bla",
			uid:         10,
			expectedRes: "root",
			wantErr:     true,
		},
		{
			name:        "root_uid_exists",
			root:        "/",
			uid:         0,
			expectedRes: "root",
			wantErr:     false,
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			username, err := lookupUsernameByUID(tt.uid, tt.root)

			if err != nil {
				if tt.wantErr {
					fmt.Println(err)
				} else {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, username)
			}

		})
	}
}

func Test_LookupGroupByUID(t *testing.T) {
	// os.Setenv("CGO_ENABLED", "0")

	uid_tests := []struct {
		name        string
		root        string
		gid         int64
		expectedRes string
		wantErr     bool
	}{
		{
			name:        "testdata_uid_exists",
			root:        "testdata",
			gid:         1,
			expectedRes: "daemon",
			wantErr:     false,
		},
		{
			name:        "testdata_uid_not_exists",
			root:        "testdata",
			gid:         10,
			expectedRes: "root",
			wantErr:     true,
		},
		{
			name:        "testdata_file_not_exists",
			root:        "testdata/bla",
			gid:         10,
			expectedRes: "root",
			wantErr:     true,
		},
		{
			name:        "root_uid_exists",
			root:        "/",
			gid:         0,
			expectedRes: "root",
			wantErr:     false,
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			groupname, err := LookupGroupnameByGID(tt.gid, tt.root)

			if err != nil {
				if tt.wantErr {
					fmt.Println(err)
				} else {
					assert.NoError(t, err)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, groupname)
			}

		})
	}

}
