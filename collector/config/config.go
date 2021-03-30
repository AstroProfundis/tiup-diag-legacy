package config

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

type ConfigCollector struct {
	Options
}

func New(opts Options) *ConfigCollector {
	return &ConfigCollector{opts}
}

func (c *ConfigCollector) Collect() error {
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
			if e := c.config(
				user.Username,
				instance.GetHost(),
				instance.GetPort(),
				instance.ComponentName(),
				instance.DeployDir()); e != nil {
				if err == nil {
					err = e
				}
			}
		}
	})

	return err
}

func (c *ConfigCollector) config(user, ip string, port int, comp, depdir string) error {
	c.GetModel().UpdateInspectionMessage(c.GetInspectionId(), fmt.Sprintf("collecting config for %s(%s:%d)...", comp, ip, port))

	if comp != "tidb" && comp != "pd" && comp != "tikv" {
		return nil
	}
	p := path.Join(c.GetHome(), "inspection", c.GetInspectionId(), "config", comp, fmt.Sprintf("%s:%d", ip, port))
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(p, comp+".toml"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command(
		"ssh",
		fmt.Sprintf("%s@%s", user, ip),
		fmt.Sprintf("cat %s/conf/%s.toml", depdir, comp),
	)
	cmd.Stdout = f
	cmd.Stderr = os.Stderr

	log.Info(cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Error("collect config file:", err)
		return err
	}

	return nil
}
