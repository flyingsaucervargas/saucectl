package flags

import (
	"encoding/csv"
	"fmt"
	"github.com/saucelabs/saucectl/internal/config"
	"strings"
)

// Simulator represents the simulator configuration.
type Simulator struct {
	config.Simulator
	Changed bool
}

// String returns a string represenation of the simulator.
func (e Simulator) String() string {
	if !e.Changed {
		return ""
	}
	return fmt.Sprintf("%+v", e.Simulator)
}

// Set sets the simulator to the values present in s.
// The input has to be a comma separated string in the format of "key=value,key2=value2".
// This method is called by cobra when CLI flags are parsed.
func (e *Simulator) Set(s string) error {
	e.Changed = true

	rec, err := csv.NewReader(strings.NewReader(s)).Read()
	if err != nil {
		return err
	}

	// TODO consider defaulting this in a common place (between config and CLI flags)
	e.PlatformName = "iOS"

	for _, v := range rec {
		vs := strings.Split(v, "=")
		val := vs[1]
		switch vs[0] {
		case "name":
			e.Name = val
		case "orientation":
			e.Orientation = val
		case "platformVersion":
			e.PlatformVersions = []string{val}
		}
	}

	return nil
}

// Type returns the value type.
func (e Simulator) Type() string {
	return "simulator"
}