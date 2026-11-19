package airuntime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoDirectProviderSDK(t *testing.T) {
	assertNoProviderSDK(t)
}

func assertNoProviderSDK(t *testing.T) {
	t.Helper()
	root := moduleRoot(t)
	banned := []string{
		"github.com/aws/aws-sdk-go",
		"github.com/aliyun/",
		"github.com/tencentyun/",
		"github.com/huaweicloud/",
		"github.com/volcengine/",
		"github.com/sashabaranov/go-openai",
		"github.com/openai/",
		"github.com/anthropics/",
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "vendor" || entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(body)
		for _, item := range banned {
			if strings.Contains(text, item) {
				t.Errorf("%s contains %s", path, item)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
