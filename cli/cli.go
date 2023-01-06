package cli

import (
	"fmt"
	"os"

	"github.com/itchyny/gojq"
	"github.com/jessevdk/go-flags"
)

type CLI struct {
	version string
}

func NewCLI(version string) *CLI {
	return &CLI{
		version: version,
	}
}

func (c *CLI) Run(args []string) int {
	err := c.run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if ex, ok := err.(Exiter); ok {
			return ex.ExitCode()
		}
		return 1
	}
	return 0
}

func (c *CLI) run(args []string) error {
	queryAndInputFiles, err := flags.ParseArgs(&options, args)
	if err != nil {
		return err
	}

	if options.Version {
		fmt.Println(c.version)
		return nil
	}

	queryString := ""
	if len(queryAndInputFiles) > 0 {
		queryString = queryAndInputFiles[0]
	}
	inputFiles := []string{}
	if len(queryAndInputFiles) > 1 {
		inputFiles = queryAndInputFiles[1:]
	}

	query, err := gojq.Parse(queryString)
	if err != nil {
		return err
	}

	iter, err := c.createInputIter(queryString, inputFiles)
	if err != nil {
		return err
	}
	defer iter.Close()

	code, err := gojq.Compile(query)
	if err != nil {
		return err
	}

	return c.process(iter, code)
}

func (c *CLI) process(iter inputIter, code *gojq.Code) error {
	var err error
	for {
		v, ok := iter.Next()
		if !ok {
			return err
		}
		if er, ok := v.(error); ok {
			c.printError(er)
			continue
		}
		if er := c.printValues(code.Run(v)); er != nil {
			c.printError(er)
		}
	}
}

func (c *CLI) printValues(iter gojq.Iter) error {
	m := c.createMarshaler()
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}

		if err := m.marshal(v, os.Stdout); err != nil {
			return err
		}
		if !options.OutputJoin {
			if options.OutputNul {
				os.Stdout.Write([]byte{'\x00'})
			} else {
				os.Stdout.Write([]byte{'\n'})
			}
		}
	}
	return nil
}

func (c *CLI) createMarshaler() marshaler {
	indent := 2
	if options.OutputCompact {
		indent = 0
	} else if options.OutputTab {
		indent = 1
	} else if i := options.OutputIndent; i != nil {
		indent = *i
	}
	f := newEncoder(options.OutputTab, indent)
	if options.OutputRaw || options.OutputJoin || options.OutputNul {
		return &rawMarshaler{f}
	}
	return f
}

func (c *CLI) printError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
}
