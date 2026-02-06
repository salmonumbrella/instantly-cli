package buildinfo

import "testing"

func TestDefaultsExist(t *testing.T) {
	_ = Version
	_ = Commit
	_ = Date
}
