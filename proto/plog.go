package proto

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
)

type Plog struct {
	// Embed the native logger
	log.Logger

	verbose bool
	writer  *os.File
	prefix  string
}

func NewLog(verbose bool) (pl *Plog, err error) {
	pl = &Plog{verbose: verbose}
	err = nil

	// For verbose output, also write to file
	if verbose {
		// Re-purpose magic tag as prefix (may want as parameter)
		pl.prefix = prefixMPS7
		var wd string
		wd, err = os.Getwd()
		if err != nil {
			err = errors.Wrap(err, "Log work dir failed")
			return
		}
		pl.writer, err = os.Create(
			filepath.Join(wd, fmt.Sprintf("%s.log", pl.prefix)))
		if err != nil {
			err = errors.Wrap(err, "Log file create failed")
			return
		}

		pl.SetFlags(log.LstdFlags)
		pl.SetPrefix(fmt.Sprintf("%s:", pl.prefix))

		// Direct log output to the file
		pl.SetOutput(pl.writer)
	}

	return
}

// Info writes message to stdout and log
func (p *Plog) Info(msg string) {
	fmt.Println(msg)

	if p.verbose {
		p.Print(msg)
	}
}

// Verbose writes message to log
func (p *Plog) Verbose(msg string) {
	// Verbose lines are skipped when
	// verbose flag is not enabled
	if p.verbose {
		p.Print(msg)
	}
}

func (p *Plog) Close() {
	if p.verbose {
		p.writer.Sync()
		p.writer.Close()
	}
}
