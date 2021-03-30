package network

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"

	"github.com/pingcap/tidb-foresight/model"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	log "github.com/sirupsen/logrus"
)

type Options interface {
	GetHome() string
	GetModel() model.Model
	GetInspectionId() string
	GetTopology() (*spec.Specification, error)
}

type NetworkCollector struct {
	Options
}

func New(opts Options) *NetworkCollector {
	return &NetworkCollector{opts}
}

func (c *NetworkCollector) Collect() error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	topo, err := c.GetTopology()
	if err != nil {
		return err
	}

	uniqueHosts := map[string]int{}
	topo.IterInstance(func(instance spec.Instance) {
		if _, found := uniqueHosts[instance.GetHost()]; !found {
			if e := c.net(user.Username, instance.GetHost()); e != nil {
				if err == nil {
					err = e
				}
			}
		}
	})

	return err
}

func (c *NetworkCollector) net(user, ip string) error {
	c.GetModel().UpdateInspectionMessage(c.GetInspectionId(), fmt.Sprintf("collecting network info for %s...", ip))
	p := path.Join(c.GetHome(), "inspection", c.GetInspectionId(), "net", ip)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(p, "ss"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command(
		"ssh",
		fmt.Sprintf("%s@%s", user, ip),
		"ss -s",
		"&&",
		"ss -lanp",
	)
	cmd.Stdout = f
	cmd.Stderr = os.Stderr

	log.Info(cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Error("get network info:", err)
		return err
	}

	return nil
}
