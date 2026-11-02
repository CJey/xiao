package xiao

import (
	"testing"
)

// 显式存入nil可遮盖父级的值，按类型取得零值；类型不符仍然panic
func TestEnvTypedGet(t *testing.T) {
	var parent = NewEnv()
	parent.Set("k", "parent")
	var child = parent.Fork()
	child.Set("k", nil)
	if child.GetString("k") != "" || child.GetAddr("k") != nil || parent.GetString("k") != "parent" {
		t.Errorf("child = %q, parent = %q", child.GetString("k"), parent.GetString("k"))
	}

	defer func() {
		if recover() == nil {
			t.Error("type mismatch should panic")
		}
	}()
	parent.GetInt("k")
}
