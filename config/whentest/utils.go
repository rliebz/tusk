// Package whentest includes constructs used in testing when clauses.
package whentest

import "github.com/rliebz/tusk/config/when"
import "github.com/rliebz/tusk/config/marshal"

// True is a when.When that always evaluates to true.
var True = when.When{}

// False is a when.When that always evaluates to false.
var False = when.When{OS: marshal.StringList{"fake"}}
