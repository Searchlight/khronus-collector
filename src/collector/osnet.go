package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/Searchlight/khronus-go-client"
	"github.com/mitchellh/mapstructure"
)

type NetworkUtilization map[string]DeviceNetworkUtilization

type DeviceNetworkUtilization struct {
	RxBytes          int64
	RxPackets        int64
	RxErrors         int64
	RxDroppedPackets int64
	TxBytes          int64
	TxPackets        int64
	TxErrors         int64
	TxDroppedPackets int64
}

func getNetStats() (*map[string]DeviceNetworkUtilization, error) {
	ret := make(map[string]DeviceNetworkUtilization)

	statFile, err := ioutil.ReadFile("/proc/net/dev")

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(statFile), "\n")
	if len(lines) <= 2 {
		return nil, fmt.Errorf("/proc/net/dev doesn't have the expected format")
	}

	for _, line := range lines[2:] {
		utilization := DeviceNetworkUtilization{}

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 16 {
			return nil, fmt.Errorf("/proc/net/dev doesn't have the expected format. Expected 16 fields found %d", len(fields))
		}
		name := strings.Trim(fields[0], ":")
		if utilization.RxBytes, err = strconv.ParseInt(fields[1], 10, 64); err != nil {
			return nil, err
		}
		if utilization.RxPackets, err = strconv.ParseInt(fields[2], 10, 64); err != nil {
			return nil, err
		}
		if utilization.RxErrors, err = strconv.ParseInt(fields[3], 10, 64); err != nil {
			return nil, err
		}
		if utilization.RxDroppedPackets, err = strconv.ParseInt(fields[4], 10, 64); err != nil {
			return nil, err
		}
		if utilization.TxBytes, err = strconv.ParseInt(fields[9], 10, 64); err != nil {
			return nil, err
		}
		if utilization.TxPackets, err = strconv.ParseInt(fields[10], 10, 64); err != nil {
			return nil, err
		}
		if utilization.TxErrors, err = strconv.ParseInt(fields[11], 10, 64); err != nil {
			return nil, err
		}
		if utilization.TxDroppedPackets, err = strconv.ParseInt(fields[12], 10, 64); err != nil {
			return nil, err
		}
		ret[name] = utilization
	}
	return &ret, nil
}

type NetCollector struct {
	host     string
	Interval int64
}

func (nc *NetCollector) Config(config map[string]interface{}) {
	fmt.Printf("%s config %#v\n", nc.Name(), nc)

	nc.host, _ = os.Hostname()

	if err := mapstructure.Decode(config, nc); err != nil {
		panic(err)
	}

	fmt.Printf("%s config %#v\n", nc.Name(), nc)
}

func (nc *NetCollector) Detect() bool {
	if runtime.GOOS == "linux" {
		return true
	} else {
		return false
	}
}

func (nc *NetCollector) Run(c chan *Metric) {

	if !nc.Detect() {
		return
	}

	for {
		pnetstats, _ := getNetStats()
		time.Sleep(time.Second * time.Duration(nc.Interval))
		netstats, _ := getNetStats()

		for k, v := range *netstats {
			c <- Gauge(nc.host + ".net." + k + ".reads.packets").Record(uint64((v.RxPackets - (*pnetstats)[k].RxPackets) / nc.Interval))
			c <- Gauge(nc.host + ".net." + k + ".reads.bytes").Record(uint64((v.RxBytes - (*pnetstats)[k].RxBytes) / nc.Interval))
			c <- Gauge(nc.host + ".net." + k + ".writes.packets").Record(uint64((v.TxPackets - (*pnetstats)[k].TxPackets) / nc.Interval))
			c <- Gauge(nc.host + ".net." + k + ".writes.bytes").Record(uint64((v.TxBytes - (*pnetstats)[k].TxBytes) / nc.Interval))
			c <- Gauge(nc.host + ".net." + k + ".errors.packets").Record(uint64((v.RxErrors + v.TxErrors) - ((*pnetstats)[k].RxErrors + (*pnetstats)[k].TxErrors)))
			c <- Gauge(nc.host + ".net." + k + ".drops.packets").Record(uint64((v.RxDroppedPackets + v.TxDroppedPackets) - ((*pnetstats)[k].RxDroppedPackets + (*pnetstats)[k].TxDroppedPackets)))
		}
	}
}

func (*NetCollector) Name() string {
	return "linux.disk.stats"
}
