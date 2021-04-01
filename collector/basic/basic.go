package basic

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pingcap/tidb-foresight/model"
	"github.com/pingcap/tidb-foresight/utils"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	log "github.com/sirupsen/logrus"
)

type Options interface {
	GetHome() string
	GetModel() model.Model
	GetInspectionId() string
	GetTopology() *spec.Specification
}

type BasicCollector struct {
	Options
}

func New(opts Options) *BasicCollector {
	return &BasicCollector{opts}
}

func (b *BasicCollector) Collect() error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	topo := b.GetTopology()
	// mutex for err
	var errMutex sync.Mutex
	var wg sync.WaitGroup

	uniqueHosts := map[string]int{}
	topo.IterInstance(func(instance spec.Instance) {
		if _, found := uniqueHosts[instance.GetHost()]; !found {
			ports := instance.UsedPorts()
			wg.Add(1)
			go func(currentHostIp string, currentPorts []int) {
				defer wg.Done()
				// collect insight on remote machine.
				if e := b.insight(user.Username, currentHostIp, ports); e != nil {
					errMutex.Lock()
					defer errMutex.Unlock()
					if err == nil {
						err = multierror.Append(err, e)
					}
				}
			}(instance.GetHost(), ports)
		}
	})
	// Note: this method thinks it will not blocked forever
	wg.Wait()

	return err
}

func (b *BasicCollector) insight(user, ip string, ports []int) error {
	b.GetModel().UpdateInspectionMessage(b.GetInspectionId(), fmt.Sprintf("collecting insight info for host %s...", ip))

	p := path.Join(b.GetHome(), "inspection", b.GetInspectionId(), "insight", ip)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(p, "collector.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	clean := exec.Command(
		"ssh",
		fmt.Sprintf("%s@%s", user, ip),
		"sudo rm -f /tmp/insight",
	)
	clean.Stdout = os.Stdout
	clean.Stderr = os.Stderr
	log.Info(clean.Args)

	install := exec.Command(
		"scp",
		path.Join(b.GetHome(), "bin", "insight"),
		fmt.Sprintf("%s@%s:/tmp/", user, ip),
	)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr

	execute := exec.Command(
		"ssh",
		fmt.Sprintf("%s@%s", user, ip),
		fmt.Sprintf("sudo chmod 755 /tmp/insight && sudo /tmp/insight --port %s", strings.Trim(strings.ReplaceAll(fmt.Sprint(ports), " ", ","), "[]")),
	)
	execute.Stdout = f
	execute.Stderr = os.Stderr

	if err := utils.RunCommands(clean, install, execute); err != nil {
		log.Error("run remote insight:", err)
		return err
	}

	return nil
}
