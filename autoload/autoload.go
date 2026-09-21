package autoload

import (
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/aileron-projects/go-properties"
)

var (
	PropsDir       = "./"
	PropsPath      = "application.properties"
	ProfilePattern = "application-${profile}.properties"
	Profile        = os.Getenv("PROPERTIES_ACTIVE_PROFILE")
)

// Props contains auto loaded properties.
var Props *properties.Properties = nil

// FileNotFound optionally handles the case of file not found.
var FileNotFound func(dir, name string) = nil

func init() {
	p := loadIfExist(PropsDir, PropsPath)
	if Profile != "" {
		path := strings.ReplaceAll(ProfilePattern, "${profile}", Profile)
		pp := loadIfExist(PropsDir, path)
		maps.Copy(p, pp)
	}
	Props = properties.NewProperties(p)
}

func loadIfExist(dir, name string) map[string]string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if notFound := FileNotFound; notFound != nil {
			notFound(dir, name)
		}
		return map[string]string{}
	}
	props, err := properties.Load(path)
	if err != nil {
		panic(err)
	}
	return props
}
