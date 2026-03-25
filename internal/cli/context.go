package cli

import (
	"fmt"
	"path/filepath"

	"example.com/dict/internal/config"
)

// Context provides shared state to all command Run() methods.
type Context struct {
	ConfigDir string
	DictName  string
	Config    config.Config
}

// ResolveDictPath returns the full path to the dictionary to use.
func (c *Context) ResolveDictPath() (string, error) {
	dictName := c.DictName
	if dictName == "" {
		dictName = c.Config.Dict
	}
	if dictName == "" {
		return "", fmt.Errorf("no dictionary configured; register one with 'dict m register <path>' and set it as default with 'dict m default <name>'")
	}
	return filepath.Join(config.DictsDir(c.ConfigDir), dictName), nil
}
