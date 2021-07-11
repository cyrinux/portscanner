package version

import (
	"fmt"
)

var (
	commit  = "0000000"
	version = "0.0.0-pre0+0000000"
	date    = "0000-00-00T00:00:00+0000"
)

// Commit returns the git commit reference
func Commit() string {
	return commit
}

// Version returns the git version, without the leading 'v'
func Version() string {
	return version
}

// Date returns the ISO-formatted date string
func Date() string {
	return date
}

// Show show the versions infos
func Show() string {
	return fmt.Sprintf("GRPCNMAPScanner v%s (%s) %s\n", Version(), Commit(), Date())
}
