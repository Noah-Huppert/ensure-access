package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Noah-Huppert/golog"
)

// fileArrayFlag allows multiple file paths to be passed as flags
type fileArrayFlag []string

// String returns a string representation of the values in an fileArrayFlag
func (f *fileArrayFlag) String() string {
	return strings.Join(*f, ",")
}

// Set adds a value to the fileArrayFlag
func (f *fileArrayFlag) Set(v string) error {
	// Check if file exists
	if _, err := os.Stat(v); os.IsNotExist(err) {
		return fmt.Errorf("file / directory \"%s\" does not exist", v)
	}

	// Add if exists
	*f = append(*f, v)

	return nil
}

// modeFlag only allows 3 digit octal numbers to be passed as flags
type modeFlag []uint32

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
		val, err := strconv.ParseUint(string(c), 8, 32)

		if err != nil {
			return err
		}

		*f = append(*f, uint32(val))
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
func newPermissions(v uint32) permissions {
	p := permissions{}

	binaryStr := strconv.FormatUint(uint64(v), 2)

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

// octal returns an octal representation of the permissions.
func (p permissions) octal() uint32 {
	var v uint32 = 0

	if p.read {
		v += 4
	}

	if p.write {
		v += 2
	}

	if p.execute {
		v += 1
	}

	return v
}

// octalString returns the value of the octal() method converted to a string
func (p permissions) octalString() string {
	return strconv.FormatUint(uint64(p.octal()), 8)
}

// or sets the values of the read, write, and execute fields by or-ing them
// with the same fields in the b argument
func (p *permissions) or(b permissions) {
	p.read = p.read || b.read
	p.write = p.write || b.write
	p.execute = p.execute || b.execute
}

// permissionsSet holds permissions for a file / directory's owner, group,
// and everyone
type permissionsSet struct {
	// owner permissions
	owner permissions

	// group permissions
	group permissions

	// everyone permissions
	everyone permissions
}

// newPermissionsSet creates a permissionsSet from an array of 3 octal digits
func newPermissionsSet(v []uint32) permissionsSet {
	p := permissionsSet{}

	p.owner = newPermissions(v[0])
	p.group = newPermissions(v[1])
	p.everyone = newPermissions(v[2])

	return p
}

// String returns a string representation of a permissionsSet
func (p permissionsSet) String() string {
	return fmt.Sprintf("%s %s %s", p.owner, p.group, p.everyone)
}

// octal returns a 3-4 digit octal representation of the permissionsSet. The
// dir argument indicates if the directory bit should be set to "1".
func (p permissionsSet) octal(dir bool) uint32 {
	var dirBit uint32 = 0
	if dir {
		dirBit = 1
	}

	n := dirBit<<9 |
		p.owner.octal()<<6 |
		p.group.octal()<<3 |
		p.everyone.octal()

	return n
}

// octalString returns a 4 digit octal representation of the permissionsSet.
// The dir argument indicates if the directory bit hsould be set to "1".
func (p permissionsSet) octalString(dir bool) string {
	v := "0"

	if dir {
		v = "1"
	}

	v += p.owner.octalString()
	v += p.group.octalString()
	v += p.everyone.octalString()

	return v
}

// or runs permissions.or() on the owner, group, and everyone fields against
// the same fields in the b argument
func (p *permissionsSet) or(b permissionsSet) {
	p.owner.or(b.owner)
	p.group.or(b.group)
	p.everyone.or(b.everyone)
}

// setPermissions ensures that every file / directory in paths has the
// permissions specified in perms
func setPermissions(logger golog.Logger, dryRun bool, paths []string,
	perms permissionsSet) error {

	for _, targetPath := range paths {
		err := filepath.Walk(targetPath, func(path string,
			info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			// Get existing permissions
			modeBits := []uint32{}
			infoMode := uint32(info.Mode())
			modeBits = append(modeBits, (0x1C0&infoMode)>>6)
			modeBits = append(modeBits, (0x38&infoMode)>>3)
			modeBits = append(modeBits, 0x7&infoMode)

			currentPerms := newPermissionsSet(modeBits)

			// Determine mode of new file
			updatedPerms := perms
			updatedPerms.or(currentPerms)

			// Determine if permissions update needs to occur
			if currentPerms != updatedPerms {
				octal := updatedPerms.octal(info.IsDir())
				fileMode := os.FileMode(octal)

				if !dryRun {
					err := os.Chmod(path, fileMode)
					if err != nil {
						return fmt.Errorf("error running chmod: %s", err.Error())
					}
				}

				dryRunStr := ""
				if dryRun {
					dryRunStr = "[dry run] "
				}
				logger.Infof("%schmod %s %s", dryRunStr, updatedPerms.octalString(info.IsDir()), path)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error ensuring permissions for \"%s\": %s", targetPath, err.Error())
		}
	}

	return nil
}

func main() {
	// {{{1 Initialize logger
	logger := golog.NewStdLogger("ensure-access")

	// {{{1 Flags
	// {{{2 Define flags
	var mode modeFlag
	var paths fileArrayFlag
	var dryRun bool

	flag.Var(&mode, "mode", "3 digit octal representation of permissions to set for files / directories")
	flag.Var(&paths, "path", "Files / directories for which permissions will be set")
	flag.BoolVar(&dryRun, "dry-run", false, "Print actions which would occur without executing actions")

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
	perms := newPermissionsSet(mode)

	// {{{1 Set permissions once on startup
	setPermissions(logger, dryRun, paths, perms)
}
