package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pingcap/tidb-foresight/collector/alert"
	"github.com/pingcap/tidb-foresight/collector/basic"
	"github.com/pingcap/tidb-foresight/collector/config"
	"github.com/pingcap/tidb-foresight/collector/dbinfo"
	"github.com/pingcap/tidb-foresight/collector/dmesg"
	logc "github.com/pingcap/tidb-foresight/collector/log"
	"github.com/pingcap/tidb-foresight/collector/metric"
	"github.com/pingcap/tidb-foresight/collector/network"
	"github.com/pingcap/tidb-foresight/model"
	"github.com/pingcap/tidb-foresight/wrapper/db"
	"github.com/pingcap/tiup/pkg/base52"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pingcap/tiup/pkg/utils/rand"
	log "github.com/sirupsen/logrus"
)

type Collector interface {
	Collect() error
}

type Options interface {
	GetHome() string
	GetInspectionId() string
	GetScrapeBegin() (time.Time, error)
	GetScrapeEnd() (time.Time, error)
}

type Manager struct {
	Options
	clusterName  string
	inspectionID string
}

func New(opts Options) Collector {
	return &Manager{
		Options:      opts,
		inspectionID: base52.Encode(time.Now().UnixNano() + rand.Int63n(1000)),
	}
}

func (m *Manager) GetInspectionId() string {
	return m.inspectionID
}

func (m *Manager) Collect() error {
	home := m.GetHome()
	start := time.Now()

	// mkdir for collection results.
	if err := os.MkdirAll(path.Join(home, "inspection", m.GetInspectionId()), os.ModePerm); err != nil {
		return err
	}

	if err := m.collectTopology(); err != nil {
		return err
	}
	if err := m.collectArgs(); err != nil {
		return err
	}
	if err := m.collectEnv(); err != nil {
		return err
	}
	if err := m.collectRemote(); err != nil {
		return err
	}

	end := time.Now()
	if cfg, err := json.Marshal(m.Options); err != nil {
		// if cannot, than panic.
		log.Error(err)
	} else {
		log.Infof(
			"Inspection %s collect with config: %s; and start from %s, ending in %s. Using time %s",
			m.GetInspectionId(), string(cfg), start.String(), end.String(), end.Sub(start).String(),
		)

	}

	return m.collectMeta(start, end)
}

// collectTopology runs in local machine.
// It move the topology file from topology/{instance_id}.json to inspection/{topology}.json
func (m *Manager) collectTopology() error {
	home := m.GetHome()

	src, err := os.Open(path.Join(home, "topology", ".json"))
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path.Join(home, "inspection", m.GetInspectionId(), "topology.json"))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// collectArgs runs in local machine.
// It generate an args.json by it's opts.
func (m *Manager) collectArgs() error {
	home := m.GetHome()

	data, err := json.Marshal(m.Options)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(home, "inspection", m.GetInspectionId(), "args.json"), data, os.ModePerm)
}

// collectArgs runs in local machine.
// It generate an args.json by it's environment variables.
func (m *Manager) collectEnv() error {
	home := m.GetHome()

	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		env[pair[0]] = os.Getenv(pair[0])
	}

	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(home, "inspection", m.GetInspectionId(), "env.json"), data, os.ModePerm)
}

func (m *Manager) collectMeta(start, end time.Time) error {
	home := m.GetHome()

	dict := map[string]time.Time{
		"create_time":  start,
		"inspect_time": start,
		"end_time":     end,
	}
	data, err := json.Marshal(dict)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(home, "inspection", m.GetInspectionId(), "meta.json"), data, os.ModePerm)
}

func (m *Manager) collectRemote() error {
	// mutex is using to protect status.
	var wg sync.WaitGroup
	var statusMutex sync.Mutex
	status := make(map[string]error)

	// build arrays for collector.
	toCollectMap := make(map[string]Collector)

	toCollectMap["alert"] = alert.New(m)
	toCollectMap["dmesg"] = dmesg.New(m)
	toCollectMap["basic"] = basic.New(m)
	toCollectMap["config"] = config.New(m)
	toCollectMap["dbinfo"] = dbinfo.New(m)
	toCollectMap["logc"] = logc.New(m)
	toCollectMap["metric"] = metric.New(m)
	//toCollectMap["profile"] = profile.New(m)
	toCollectMap["network"] = network.New(m)

	for item, collector := range toCollectMap {
		wg.Add(1)
		go func(innerCollector Collector, key string) {
			defer wg.Done()
			collected := innerCollector.Collect()
			log.Info(fmt.Sprintf("[Collector] Inspection collect %s done.", key))
			statusMutex.Lock()
			defer statusMutex.Unlock()
			status[key] = collected
		}(collector, item)
	}

	wg.Wait()
	return m.collectStatus(status)
}

func (m *Manager) collectStatus(status map[string]error) error {
	home := m.GetHome()

	dict := make(map[string]map[string]string)
	for item, err := range status {
		if err == nil {
			dict[item] = map[string]string{
				"status": "success",
			}
		} else {
			dict[item] = map[string]string{
				"status":  "error",
				"message": err.Error(),
			}
		}
	}

	data, err := json.Marshal(dict)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(home, "inspection", m.GetInspectionId(), "status.json"), data, os.ModePerm)
}

func (m *Manager) GetTopology() (*spec.Specification, error) {
	metadata, err := spec.ClusterMetadata(m.clusterName)
	if err != nil {
		return nil, err
	}

	return metadata.Topology, nil
}

func (m *Manager) GetPrometheusEndpoint() (string, error) {
	topo, err := m.GetTopology()
	if err != nil {
		return "", err
	}

	for _, host := range topo.Monitors {
		return fmt.Sprintf("%s:%d", host.Host, host.Port), nil
	}

	return "", errors.New("component prometheus not found")
}

func (m *Manager) GetTidbStatusEndpoints() ([]string, error) {
	endpoints := []string{}

	topo, err := m.GetTopology()
	if err != nil {
		return endpoints, err
	}

	for _, host := range topo.TiDBServers {
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", host.Host, host.StatusPort))
	}

	return endpoints, nil
}

func (m *Manager) GetModel() model.Model {
	db, err := db.Open(path.Join(m.GetHome(), "sqlite.db"))
	if err != nil {
		log.Panic(err)
	}

	return model.New(db)
}
