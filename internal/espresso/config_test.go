package espresso

import (
	"errors"
	"github.com/saucelabs/saucectl/internal/config"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestValidateThrowsErrors(t *testing.T) {
	testCases := []struct {
		name        string
		p           *Project
		expectedErr error
	}{
		{
			name:        "validating throws error on empty app",
			p:           &Project{},
			expectedErr: errors.New("missing path to app .apk"),
		},
		{
			name: "validating throws error on app missing .apk",
			p: &Project{
				Espresso: Espresso{
					App: "/path/to/app",
				},
			},
			expectedErr: errors.New("invaild application file: /path/to/app, make sure extension is .apk"),
		},
		{
			name: "validating throws error on empty app",
			p: &Project{
				Espresso: Espresso{
					App: "/path/to/app.apk",
				},
			},
			expectedErr: errors.New("missing path to test app .apk"),
		},
		{
			name: "validating throws error on test app missing .apk",
			p: &Project{
				Espresso: Espresso{
					App:     "/path/to/app.apk",
					TestApp: "/path/to/testApp",
				},
			},
			expectedErr: errors.New("invaild test application file: /path/to/testApp, make sure extension is .apk"),
		},
		{
			name: "validating throws error on missing suites",
			p: &Project{
				Espresso: Espresso{
					App:     "/path/to/app.apk",
					TestApp: "/path/to/testApp.apk",
				},
			},
			expectedErr: errors.New("no suites defined"),
		},
		{
			name: "validating throws error on missing devices",
			p: &Project{
				Espresso: Espresso{
					App:     "/path/to/app.apk",
					TestApp: "/path/to/testApp.apk",
				},
				Suites: []Suite{
					{
						Name:    "no devices",
						Devices: []config.Device{},
					},
				},
			},
			expectedErr: errors.New("missing devices or emulators configuration for suite: no devices"),
		},
		{
			name: "validating throws error on missing device name",
			p: &Project{
				Espresso: Espresso{
					App:     "/path/to/app.apk",
					TestApp: "/path/to/testApp.apk",
				},
				Suites: []Suite{
					{
						Name: "empty emulator name",
						Emulators: []config.Emulator{
							{
								Name: "",
							},
						},
					},
				},
			},
			expectedErr: errors.New("missing emulator name for suite: empty emulator name. Emulators index: 0"),
		},
		{
			name: "validating throws error on missing Emulator suffix on device name",
			p: &Project{
				Espresso: Espresso{
					App:     "/path/to/app.apk",
					TestApp: "/path/to/testApp.apk",
				},
				Suites: []Suite{
					{
						Name: "no emulator device name",
						Emulators: []config.Emulator{
							{
								Name: "Android GoogleApi something",
							},
						},
					},
				},
			},
			expectedErr: errors.New("missing `emulator` in emulator name: Android GoogleApi something. Suite name: no emulator device name. Emulators index: 0"),
		},
		{
			name: "validating throws error on missing platform versions",
			p: &Project{
				Espresso: Espresso{
					App:     "/path/to/app.apk",
					TestApp: "/path/to/testApp.apk",
				},
				Suites: []Suite{
					{
						Name: "no emulator device name",
						Emulators: []config.Emulator{
							{
								Name:             "Android GoogleApi Emulator",
								PlatformVersions: []string{},
							},
						},
					},
				},
			},
			expectedErr: errors.New("missing platform versions for emulator: Android GoogleApi Emulator. Suite name: no emulator device name. Emulators index: 0"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(*tc.p)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedErr.Error(), err.Error())
		})
	}
}

func TestFromFile(t *testing.T) {
	dir := fs.NewDir(t, "espresso-cfg",
		fs.WithFile("config.yml", `apiVersion: v1alpha
kind: espresso
espresso:
  app: ./tests/apps/calc.apk
  testApp: ./tests/apps/calc-success.apk
suites:
  - name: "saucy barista"
    devices:
      - name: "Device name"
        platformVersion: 8.1
        options:
          deviceType: TABLET
    emulators:
      - name: "Google Pixel C GoogleAPI Emulator"
        platformVersions:
          - "8.1"
`, fs.WithMode(0655)))
	defer dir.Remove()

	cfg, err := FromFile(filepath.Join(dir.Path(), "config.yml"))
	if err != nil {
		t.Errorf("expected error: %v, got: %v", nil, err)
	}
	expected := Project{
		ConfigFilePath: filepath.Join(dir.Path(), "config.yml"),
		Espresso: Espresso{
			App:     "./tests/apps/calc.apk",
			TestApp: "./tests/apps/calc-success.apk",
		},
		Suites: []Suite{
			{
				Name: "saucy barista",
				Devices: []config.Device{
					{
						Name:            "Device name",
						PlatformVersion: "8.1",
						Options: config.DeviceOptions{
							DeviceType: "TABLET",
						},
					},
				},
				Emulators: []config.Emulator{
					{
						Name:         "Google Pixel C GoogleAPI Emulator",
						PlatformVersions: []string{
							"8.1",
						},
					},
				}},
		},
	}
	if !reflect.DeepEqual(cfg.Espresso, expected.Espresso) {
		t.Errorf("expected: %v, got: %v", expected.Espresso, cfg.Espresso)
	}
	if !reflect.DeepEqual(cfg.Suites, expected.Suites) {
		t.Errorf("expected: %v, got: %v", expected.Suites, cfg.Suites)
	}

}
