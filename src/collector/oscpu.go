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

type CpuStat struct {
	Name      string // CpuName or Total
	User      uint64 // Normal processes executing in user mode
	Nice      uint64 // Niced processes executing in user mode
	Sys       uint64 // Processes executing in kernel mode
	Idle      uint64 // Twiddling thumbs
	Wait      uint64 // Waiting for I/O to complete
	Irq       uint64 // Servicing interrupts
	SoftIrq   uint64 // Servicing softirqs
	Stolen    uint64 // Ticks spent executing other virtual hosts
	Guest     uint64 // Time spent running a virtual CPU for guest operating systems
	GuestNice uint64 // Time spent running a niced CPU for guest operating systems
}

type CpuStats struct {
	Total       CpuStat
	Cpus        []CpuStat
	Intr        []uint64
	Ctxt        uint64
	BootTime    uint64
	Processes   uint64
	ProcRunning uint64
	ProcBlocked uint64
	SoftIrq     []uint64
}

func parseCpu(cpu *CpuStat, fields []string) error {
	cpu.Name = fields[0]

	var err error

	if cpu.User, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
		return err
	}
	if cpu.Nice, err = strconv.ParseUint(fields[2], 10, 64); err != nil {
		return err
	}
	if cpu.Sys, err = strconv.ParseUint(fields[3], 10, 64); err != nil {
		return err
	}
	if cpu.Idle, err = strconv.ParseUint(fields[4], 10, 64); err != nil {
		return err
	}
	if cpu.Wait, err = strconv.ParseUint(fields[5], 10, 64); err != nil {
		return err
	}
	if cpu.Irq, err = strconv.ParseUint(fields[6], 10, 64); err != nil {
		return err
	}
	if cpu.SoftIrq, err = strconv.ParseUint(fields[7], 10, 64); err != nil {
		return err
	}
	if cpu.Stolen, err = strconv.ParseUint(fields[8], 10, 64); err != nil {
		return err
	}
	if cpu.Guest, err = strconv.ParseUint(fields[9], 10, 64); err != nil {
		return err
	}
	if cpu.GuestNice, err = strconv.ParseUint(fields[10], 10, 64); err != nil {
		return err
	}

	return nil
}

type LoadAverage struct {
	One        float64
	Five       float64
	Fifteen    float64
	Runable    uint64
	Scheduling uint64
	LastPid    uint64
}

func getLoadAverage() (*LoadAverage, error) {
	loadStats := "/proc/loadavg"
	ret := LoadAverage{}

	if _, err := os.Stat(loadStats); err != nil {
		return nil, fmt.Errorf("%s not exists", loadStats)
	}

	contents, err := ioutil.ReadFile(loadStats)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))
	line, _, err := reader.ReadLine()

	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(line))

	if ret.One, err = strconv.ParseFloat(fields[0], 64); err != nil {
		return nil, err
	}

	if ret.Five, err = strconv.ParseFloat(fields[1], 64); err != nil {
		return nil, err
	}

	if ret.Fifteen, err = strconv.ParseFloat(fields[2], 64); err != nil {
		return nil, err
	}

	t := strings.Split(fields[3], "/")

	if ret.Runable, err = strconv.ParseUint(t[0], 10, 64); err != nil {
		return nil, err
	}

	if ret.Scheduling, err = strconv.ParseUint(t[1], 10, 64); err != nil {
		return nil, err
	}

	if ret.LastPid, err = strconv.ParseUint(fields[4], 10, 64); err != nil {
		return nil, err
	}

	return &ret, nil
}

func getCpuStats() (*CpuStats, error) {
	procStats := "/proc/stat"
	ret := CpuStats{}

	if _, err := os.Stat(procStats); err != nil {
		return nil, fmt.Errorf("%s not exists", procStats)
	}

	contents, err := ioutil.ReadFile(procStats)
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

		if string(line[0:3]) == "cpu" {
			cpu := CpuStat{}
			err := parseCpu(&cpu, fields)
			if err != nil {
				return nil, err
			}
			if string(line[3]) == " " {
				ret.Total = cpu
			} else {
				ret.Cpus = append(ret.Cpus, cpu)
			}
		} else {
			switch fields[0] {
			case "intr":
				continue
				for _, v := range fields[1:] {
					uif, err := strconv.ParseUint(v, 10, 64)
					if err != nil {
						return nil, err
					}
					ret.Intr = append(ret.Intr, uif)
				}
			case "ctxt":
				continue
				ret.Ctxt, err = strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
			case "btime":
				continue
				ret.BootTime, err = strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
			case "processes":
				continue
				ret.Processes, err = strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
			case "procs_running":
				continue
				ret.ProcRunning, err = strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
			case "procs_blocked":
				continue
				ret.ProcBlocked, err = strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
			case "softirq":
				continue
				for _, v := range fields[1:] {
					uif, err := strconv.ParseUint(v, 10, 64)
					if err != nil {
						return nil, err
					}
					ret.SoftIrq = append(ret.SoftIrq, uif)
				}
			}
		}
	}
	return &ret, nil
}

type CpuCollector struct {
	cpuload  LoadAverage
	cpustats CpuStats
	host     string
	Interval int64
}

func (cc *CpuCollector) Config(config map[string]interface{}) {
	fmt.Printf("%s config %#v\n", cc.Name(), cc)

	cc.host, _ = os.Hostname()

	if err := mapstructure.Decode(config, cc); err != nil {
		panic(err)
	}

}

func (cc *CpuCollector) Detect() bool {
	if runtime.GOOS == "linux" {
		return true
	} else {
		return false
	}
}

func (cc *CpuCollector) Run(c chan *Metric) {

	// [TODO] Handler Error

	ptotal, err := getCpuStats()

	if err != nil {
		fmt.Errorf("%s\n", err)
	}

	time.Sleep(time.Duration(cc.Interval) * time.Second)

	for {
		total, err := getCpuStats()
		if err != nil {
			fmt.Errorf("%s\n", err)
		}
		cpuload, err := getLoadAverage()
		if err != nil {
			fmt.Errorf("%s\n", err)
		}

		c <- Gauge(cc.host + ".cpu_load.one").Record(uint64(cpuload.One * 100))
		c <- Gauge(cc.host + ".cpu_load.five").Record(uint64(cpuload.Five * 100))
		c <- Gauge(cc.host + ".cpu_load.fifteen").Record(uint64(cpuload.Fifteen * 100))

		tc := float64((total.Total.Idle - ptotal.Total.Idle) +
			(total.Total.Irq - ptotal.Total.Irq) +
			(total.Total.SoftIrq - ptotal.Total.SoftIrq) +
			(total.Total.Stolen - ptotal.Total.Stolen) +
			(total.Total.Sys - ptotal.Total.Sys) +
			(total.Total.User - ptotal.Total.User) +
			(total.Total.Nice - ptotal.Total.Nice) +
			(total.Total.Wait - ptotal.Total.Wait))

		c <- Gauge(cc.host + ".cpu_total.idle").Record(uint64(float64(total.Total.Idle-ptotal.Total.Idle) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.irq").Record(uint64(float64(total.Total.Irq-ptotal.Total.Irq) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.softirq").Record(uint64(float64(total.Total.SoftIrq-ptotal.Total.SoftIrq) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.stolen").Record(uint64(float64(total.Total.Stolen-ptotal.Total.Stolen) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.sys").Record(uint64(float64(total.Total.Sys-ptotal.Total.Sys) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.user").Record(uint64(float64(total.Total.User-ptotal.Total.User) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.nice").Record(uint64(float64(total.Total.Nice-ptotal.Total.Nice) * 100 / tc))
		c <- Gauge(cc.host + ".cpu_total.wait").Record(uint64(float64(total.Total.Wait-ptotal.Total.Wait) * 100 / tc))

		ptotal = total

		time.Sleep(time.Duration(cc.Interval) * time.Second)
	}
}

func (*CpuCollector) Name() string {
	return "linux.cpu.stats"
}
