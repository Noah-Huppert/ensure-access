package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/Noah-Huppert/golog"
)

// arrayFlag allows a command line flag to be passed multiple times
type arrayFlag []string

// String returns a string representation of the values in an arrayFlag
func (f *arrayFlag) String() string {
	return strings.Join(*f, ",")
}

// Set adds a value to the arrayFlag
func (f *arrayFlag) Set(v string) error {
	*f = append(*f, v)
	return nil
}

// modeFlag only allows 3 digit octal numbers to be passed as flags
type modeFlag []int64

// String returns a string representation of the modeFlag
func (f *modeFlag) String() string {
	s := ""

	for _, v := range *f {
		s += string(v)
	}

	return s
}

// Set the value of a modeFlag
func (f *modeFlag) Set(v string) error {
	var err = errors.New("must be 3 octal digits")

	// Check 3 digits provided
	if len(v) != 3 {
		return err
	}

	// Check each character is a valid octal digit
	for _, c := range v {
		val, err := strconv.ParseInt(string(c), 8, 32)

		if err != nil {
			return err
		}

		*f = append(*f, val)
	}

	return nil
}

// permissions indicates which file operations a user / group can perform
type permissions struct {
	// read operation
	read bool

	// write operation
	write bool

	// execute operation
	execute bool
}

// newPermissions creates a permissions struct from a single octal digit
func newPermissions(v int64) permissions {
	p := permissions{}

	binaryStr := strconv.FormatInt(v, 2)

	for len(binaryStr) < 3 {
		binaryStr = fmt.Sprintf("0%s", binaryStr)
	}

	if binaryStr[0] == '1' {
		p.read = true
	}

	if binaryStr[1] == '1' {
		p.write = true
	}

	if binaryStr[2] == '1' {
		p.execute = true
	}

	return p
}

// String returns a string representation of a permissions struct
func (p permissions) String() string {
	s := ""

	if p.read {
		s += "r"
	}

	if p.write {
		s += "w"
	}

	if p.execute {
		s += "x"
	}

	return s
}

func main() {
	// {{{1 Initialize logger
	logger := golog.NewStdLogger("ensure-access")

	// {{{1 Flags
	// {{{2 Define flags
	var mode modeFlag
	var paths arrayFlag

	flag.Var(&mode, "mode", "3 digit octal representation of permissions "+
		"to set for files / directories")
	flag.Var(&paths, "path", "Files / directories for which permissions "+
		"will be set")

	// {{{2 Parse flags
	flag.Parse()

	// {{{2 Verify
	// {{{3 mode
	if len(mode) == 0 {
		logger.Fatal("-mode option required")
	}

	// {{{3 paths
	if len(paths) == 0 {
		logger.Fatal("-path option required")
	}

	// {{{1 Extract permissions from mode octal
	ownerPerms := newPermissions(mode[0])
	groupPerms := newPermissions(mode[1])
	everyonePerms := newPermissions(mode[2])

	logger.Debugf("%s %s %s", ownerPerms, groupPerms, everyonePerms)
}
