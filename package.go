package flock

import (
	"github.com/FortifiedCode/flock/internal/processor"
	"github.com/FortifiedCode/plover"
	"os"
	"path/filepath"
)

const defaultConfigName = "processor0"

type Processor struct {
	value *processor.Processor
}

func New(repository *plover.Repository) Processor {

	ex, err := os.Executable()
	if err != nil {
		panic("failed to get the working directory of the executable")
	}

	workingDir := filepath.Dir(ex)

	paths := [4]string{
		filepath.Join(workingDir, "processor.toml"),
		filepath.Join(workingDir, "..", "processor.toml"),
		filepath.Join(workingDir, "..", "..", "processor.toml"),
		filepath.Join(workingDir, "..", "..", "processor.toml"),
	}

	var cfg *processor.Config = nil
	for _, path := range paths {
		cfg, err = processor.Load(path)
		if err == nil {
			break
		}
	}

	// there 'err' is non nil when we could not find a config
	if err != nil {
		cfg = processor.NewConfig(defaultConfigName)
	}

	p := processor.New(cfg, repository)
	return Processor{value: p}
}

func (p Processor) Connect() {

	p.value.Connect()
}
