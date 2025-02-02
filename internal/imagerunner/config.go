package imagerunner

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/saucelabs/saucectl/internal/config"
	"github.com/saucelabs/saucectl/internal/msg"
	"github.com/saucelabs/saucectl/internal/region"
)

var (
	Kind       = "imagerunner"
	APIVersion = "v1alpha"

	ValidWorkloadType = []string{
		"webdriver",
		"other",
	}
)

type Project struct {
	config.TypeDef `yaml:",inline" mapstructure:",squash"`
	Defaults       Defaults           `yaml:"defaults" json:"defaults"`
	Sauce          config.SauceConfig `yaml:"sauce,omitempty" json:"sauce"` // The only field that's used within 'sauce' is region.
	Suites         []Suite            `yaml:"suites,omitempty" json:"suites"`
	Artifacts      config.Artifacts   `yaml:"artifacts,omitempty" json:"artifacts"`
}

type Defaults struct {
	Suite `yaml:",inline" mapstructure:",squash"`
}

type Suite struct {
	Name          string            `yaml:"name,omitempty" json:"name"`
	Image         string            `yaml:"image,omitempty" json:"image"`
	ImagePullAuth ImagePullAuth     `yaml:"imagePullAuth,omitempty" json:"imagePullAuth"`
	EntryPoint    string            `yaml:"entrypoint,omitempty" json:"entrypoint"`
	Files         []File            `yaml:"files,omitempty" json:"files"`
	Artifacts     []string          `yaml:"artifacts,omitempty" json:"artifacts"`
	Env           map[string]string `yaml:"env,omitempty" json:"env"`
	Timeout       time.Duration     `yaml:"timeout,omitempty" json:"timeout"`
	Workload      string            `yaml:"workload,omitempty" json:"workload,omitempty"`
}

type ImagePullAuth struct {
	User  string `yaml:"user,omitempty" json:"user"`
	Token string `yaml:"token,omitempty" json:"token"`
}

type File struct {
	Src string `yaml:"src,omitempty" json:"src"`
	Dst string `yaml:"dst,omitempty" json:"dst"`
}

func FromFile(cfgPath string) (Project, error) {
	var p Project

	if err := config.Unmarshal(cfgPath, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SetDefaults applies config defaults in case the user has left them blank.
func SetDefaults(p *Project) {
	if p.Kind == "" {
		p.Kind = Kind
	}

	if p.APIVersion == "" {
		p.APIVersion = APIVersion
	}

	if p.Sauce.Concurrency < 1 {
		p.Sauce.Concurrency = 1
	}

	if p.Sauce.Concurrency > 5 {
		log.Warn().Msgf(msg.ImageRunnerMaxConcurrency, p.Sauce.Concurrency)
		p.Sauce.Concurrency = 5
	}

	if p.Defaults.Timeout < 0 {
		p.Defaults.Timeout = 0
	}

	p.Sauce.Tunnel.SetDefaults()
	p.Sauce.Metadata.SetDefaultBuild()

	for i, suite := range p.Suites {
		if suite.Timeout <= 0 {
			p.Suites[i].Timeout = p.Defaults.Timeout
		}

		if suite.Workload == "" {
			p.Suites[i].Workload = p.Defaults.Workload
		}
	}
}

func Validate(p Project) error {
	regio := region.FromString(p.Sauce.Region)
	if regio == region.None {
		return errors.New(msg.MissingRegion)
	}

	if len(p.Suites) == 0 {
		return errors.New(msg.EmptySuite)
	}

	for _, suite := range p.Suites {
		if suite.Workload == "" {
			return fmt.Errorf(msg.MissingImageRunnerWorkloadType, suite.Name)
		}

		if !sliceContainsString(ValidWorkloadType, suite.Workload) {
			return fmt.Errorf(msg.InvalidImageRunnerWorkloadType, suite.Workload, suite.Name)
		}

		if suite.Image == "" {
			return fmt.Errorf(msg.MissingImageRunnerImage, suite.Name)
		}
	}
	return nil
}

func sliceContainsString(slice []string, val string) bool {
	for _, value := range slice {
		if value == val {
			return true
		}
	}
	return false
}

// FilterSuites filters out suites in the project that don't match the given suite name.
func FilterSuites(p *Project, suiteName string) error {
	for _, s := range p.Suites {
		if s.Name == suiteName {
			p.Suites = []Suite{s}
			return nil
		}
	}
	return fmt.Errorf(msg.SuiteNameNotFound, suiteName)
}
