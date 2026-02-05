package cli

import (
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
func (c *Context) ResolveDictPath() string {
	dictName := c.DictName
	if dictName == "" {
		dictName = c.Config.Dict
	}
	return filepath.Join(config.DictsDir(c.ConfigDir), dictName)
}
