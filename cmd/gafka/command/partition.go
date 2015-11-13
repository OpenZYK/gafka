package command

import (
	"strings"

	"github.com/funkygao/gocli"
)

type Partition struct {
	Ui cli.Ui
}

func (this *Partition) Run(args []string) (exitCode int) {
	return

}

func (*Partition) Synopsis() string {
	return "Add partition num to a topic TODO"
}

func (*Partition) Help() string {
	help := `
Usage: gafka partition -z zone -c cluster -t topic [options]
`
	return strings.TrimSpace(help)
}