package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultDebugLogFile = "debug.log"

// debugFlag is the value of the --debug flag. A bare --debug (or a boolean
// value such as --debug=true) toggles debug logging to the default file,
// any other value enables debug logging and is used as the log location.
type debugFlag struct {
	enabled bool
	path    string
}

func (d *debugFlag) String() string {
	if d == nil || !d.enabled {
		return ""
	}
	return d.path
}

func (d *debugFlag) Set(s string) error {
	if b, err := strconv.ParseBool(s); err == nil {
		d.enabled = b
		d.path = ""
		return nil
	}
	if s == "" {
		return errors.New("empty debug log path")
	}
	d.enabled = true
	d.path = s
	return nil
}

// IsBoolFlag allows the flag to be passed without a value (--debug).
func (d *debugFlag) IsBoolFlag() bool {
	return true
}

// resolveDebugLogPath returns the file debug logs are written to. An empty
// path resolves to debug.log in the working directory, a directory (an
// existing one, or any path ending in a separator) resolves to debug.log
// inside it and anything else is used as the file path as is.
func resolveDebugLogPath(p string) string {
	if p == "" {
		return defaultDebugLogFile
	}
	if strings.HasSuffix(p, string(os.PathSeparator)) || strings.HasSuffix(p, "/") {
		return filepath.Join(p, defaultDebugLogFile)
	}
	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
		return filepath.Join(p, defaultDebugLogFile)
	}
	return p
}
