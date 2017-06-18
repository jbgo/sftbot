package command

import (
	"bytes"
	"flag"
)

type Command interface {
	InitFlags() *flag.FlagSet
}

func helpOptions(c Command) string {
	flags := c.InitFlags()

	options := ""
	buf := bytes.NewBufferString(options)
	flags.SetOutput(buf)
	flags.PrintDefaults()

	return "Options:\n\n" + buf.String()
}
