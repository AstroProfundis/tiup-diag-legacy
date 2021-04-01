package dmesg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pingcap/tidb-foresight/model"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/cluster/executor"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pingcap/tiup/pkg/cluster/task"
	log "github.com/sirupsen/logrus"
)

type Options interface {
	GetHome() string
	GetModel() model.Model
	GetInspectionId() string
	GetTopology() *spec.Specification
	GetUser() string
	GetPasswd() string
	GetIdentityFile() string
}

type DmesgCollector struct {
	Options
}

func New(opts Options) *DmesgCollector {
	return &DmesgCollector{opts}
}

func (c *DmesgCollector) Collect() error {
	topo := c.GetTopology()
	var ctask []*task.StepDisplay

	uniqueHosts := map[string]int{}
	topo.IterInstance(func(instance spec.Instance) {
		if _, found := uniqueHosts[instance.GetHost()]; !found {
			uniqueHosts[instance.GetHost()] = instance.GetSSHPort()
			t := task.NewBuilder().
				RootSSH(
					instance.GetHost(),
					instance.GetSSHPort(),
					c.GetUser(),
					c.GetPasswd(),
					c.GetIdentityFile(),
					"",
					30,
					executor.SSHTypeBuiltin,
					executor.SSHTypeBuiltin,
				).
				Shell(
					instance.GetHost(),
					"dmesg",
					"",
					true,
				)
			ctask = append(ctask, t.BuildAsStep(fmt.Sprintf("collecting dmesg from %s", instance.GetHost())))
		}
	})

	t := task.NewBuilder().
		ParallelStep("Collect kernel messages", false, ctask...).
		Build()

	ctx := ctxt.New(context.Background())
	if err := t.Execute(ctx); err != nil {
		return err
	}
	return nil
}

func (c *DmesgCollector) dmesg(user, ip string) error {
	c.GetModel().UpdateInspectionMessage(c.GetInspectionId(), fmt.Sprintf("collecting dmesg info for %s...", ip))

	p := path.Join(c.GetHome(), "inspection", c.GetInspectionId(), "dmesg", ip)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(p, "dmesg"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command(
		"ssh",
		fmt.Sprintf("%s@%s", user, ip),
		"sudo dmesg",
	)
	cmd.Stdout = f
	cmd.Stderr = os.Stderr

	log.Info(cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Error("get dmesg info:", err)
		return err
	}

	return nil
}
