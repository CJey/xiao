package xiao

import (
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func TestReplaceLoggerConcurrent(t *testing.T) {
	t.Cleanup(ReplaceLogger(zap.NewNop()))
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			ReplaceLogger(zap.NewNop())
		}
	})
	wg.Go(func() {
		for range 100 {
			NamedContext("race").Info("x")
		}
	})
	wg.Wait()
}

// 文件路径与file://形式的URL同样自动建目录，且不在当前目录留下杂物
func TestSimpleLoggerOutpath(t *testing.T) {
	var dir = t.TempDir()
	t.Chdir(dir)
	for _, out := range []string{
		filepath.Join(dir, "abs", "a.log"),
		"file://" + filepath.Join(dir, "url", "a.log"),
		filepath.Join("rel", "a.log"),
	} {
		var z, err = NewSimpleLogger("info", out, "json", false)
		if err != nil {
			t.Errorf("%s: %v", out, err)
			continue
		}
		z.Info("x")
	}

	var names []string
	var entries, _ = os.ReadDir(dir)
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if !slices.Equal(names, []string{"abs", "rel", "url"}) {
		t.Errorf("entries = %v", names)
	}
}
