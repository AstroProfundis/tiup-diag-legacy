module github.com/pingcap/tidb-foresight

go 1.16

require (
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/fatih/color v1.12.0
	github.com/fatih/structs v1.1.0
	github.com/google/uuid v1.2.0
	github.com/influxdata/influxdb v1.9.1
	github.com/joomcode/errorx v1.0.3
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pingcap/check v0.0.0-20200212061837-5e12011dc712
	github.com/pingcap/errors v0.11.5-0.20201126102027-b0a155152ca3
	github.com/pingcap/log v0.0.0-20201112100606-8f1e84a3abc8 // indirect
	github.com/pingcap/tidb-insight v0.4.0-dev.1.0.20210610034452-7985a5287b50 // indirect
	github.com/pingcap/tidb-insight/collector v0.0.0-20210610034452-7985a5287b50
	github.com/pingcap/tiup v1.6.0-dev.0.20210802034506-35abe88f2cc1
	github.com/prometheus/common v0.29.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.17.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/appleboy/easyssh-proxy => github.com/AstroProfundis/easyssh-proxy v1.3.10-0.20210615044136-d52fc631316d
