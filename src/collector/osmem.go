package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/Searchlight/khronus-go-client"
	"github.com/mitchellh/mapstructure"
)

type MemCollector struct {
	host     string
	Interval int64
}

type MemStat struct {
	MemTotal          uint64
	MemFree           uint64
	MemAvailable      uint64
	Buffers           uint64
	Cached            uint64
	SwapCached        uint64
	Active            uint64
	Inactive          uint64
	ActiveAnon        uint64
	InactiveAnon      uint64
	ActiveFile        uint64
	InactiveFile      uint64
	Unevictable       uint64
	Mlocked           uint64
	SwapTotal         uint64
	SwapFree          uint64
	Dirty             uint64
	Writeback         uint64
	AnonPages         uint64
	Mapped            uint64
	Shmem             uint64
	Slab              uint64
	SReclaimable      uint64
	SUnreclaim        uint64
	KernelStack       uint64
	PageTables        uint64
	NFS_Unstable      uint64
	Bounce            uint64
	WritebackTmp      uint64
	CommitLimit       uint64
	CommittedAs       uint64
	VmallocTotal      uint64
	VmallocUsed       uint64
	VmallocChunk      uint64
	HardwareCorrupted uint64
	AnonHugePages     uint64
}

func getMem() (*MemStat, error) {
	memStats := "/proc/meminfo"
	ret := MemStat{}

	if _, err := os.Stat(memStats); err != nil {
		return nil, fmt.Errorf("%s not exists", memStats)
	}

	contents, err := ioutil.ReadFile(memStats)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))

	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		}

		fields := strings.Fields(string(line))

		switch fields[0] {
		case "Active:":
			if ret.Active, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "ActiveAnon:":
			if ret.ActiveAnon, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "ActiveFile:":
			if ret.ActiveFile, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "AnonHugePages:":
			if ret.AnonHugePages, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "AnonPages:":
			if ret.AnonHugePages, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Bounce:":
			if ret.Bounce, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Buffers:":
			if ret.Buffers, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Cached:":
			if ret.Cached, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "CommitLimit:":
			if ret.CommitLimit, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "CommittedAs:":
			if ret.CommittedAs, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Dirtys:":
			if ret.CommittedAs, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "HardwareCorrupted:":
			if ret.HardwareCorrupted, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Inactive:":
			if ret.Inactive, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "InactiveAnon:":
			if ret.InactiveAnon, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "InactiveFile:":
			if ret.InactiveFile, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "KernelStack:":
			if ret.KernelStack, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Mapped:":
			if ret.Mapped, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "MemAvailable:":
			if ret.MemAvailable, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "MemFree:":
			if ret.MemFree, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "MemTotal:":
			if ret.MemTotal, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Mlocked:":
			if ret.Mlocked, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "NFS_Unstable:":
			if ret.NFS_Unstable, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "PageTables:":
			if ret.PageTables, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "SReclaimable:":
			if ret.SReclaimable, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "SUnreclaim:":
			if ret.SUnreclaim, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Shmem:":
			if ret.Shmem, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Slab:":
			if ret.Slab, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "SwapCached:":
			if ret.SwapCached, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "SwapFree:":
			if ret.SwapFree, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "SwapTotal:":
			if ret.SwapTotal, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Unevictable:":
			if ret.Unevictable, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "VmallocChunk:":
			if ret.VmallocChunk, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "VmallocTotal:":
			if ret.VmallocTotal, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "VmallocUsed:":
			if ret.VmallocUsed, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "Writeback:":
			if ret.Writeback, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		case "WritebackTmp:":
			if ret.WritebackTmp, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
				return nil, err
			}
		}
	}

	return &ret, nil

}

func (mc *MemCollector) Config(config map[string]interface{}) {
	fmt.Printf("%s config %#v\n", mc.Name(), mc)

	mc.host, _ = os.Hostname()

	if err := mapstructure.Decode(config, mc); err != nil {
		panic(err)
	}

	fmt.Printf("%s config %#v\n", mc.Name(), mc)
}

func (mc *MemCollector) Detect() bool {
	if runtime.GOOS == "linux" {
		return true
	} else {
		return false
	}
}

func (mc *MemCollector) Run(c chan *Metric) {

	if !mc.Detect() {
		return
	}

	for {
		mem, err := getMem()

		if err != nil {
			fmt.Println(err)
		}

		select {
		case <-time.After(time.Second * time.Duration(mc.Interval)):
			c <- Gauge(mc.host + ".mem.cached").Record(mem.Cached)
			c <- Gauge(mc.host + ".mem.buffer").Record(mem.Buffers)
			c <- Gauge(mc.host + ".mem.free").Record(mem.MemFree)
			c <- Gauge(mc.host + ".mem.used").Record(mem.MemTotal - mem.MemFree - mem.Cached - mem.Buffers)
		}
	}
}

func (*MemCollector) Name() string {
	return "linux.mem.stats"
}
