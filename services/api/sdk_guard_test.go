package sdkguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoDirectProviderSDK(t *testing.T) {
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
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
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
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
