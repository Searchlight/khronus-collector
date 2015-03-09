package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/despegar/khronus-go-client"
	"github.com/mitchellh/mapstructure"
)

type DiskStats struct {
	Major             int
	Minor             int
	Device            string
	ReadRequests      uint64 // Total number of reads completed successfully.
	ReadMerged        uint64 // Adjacent read requests merged in a single req.
	ReadSectors       uint64 // Total number of sectors read successfully.
	MsecRead          uint64 // Total number of ms spent by all reads.
	WriteRequests     uint64 // total number of writes completed successfully.
	WriteMerged       uint64 // Adjacent write requests merged in a single req.
	WriteSectors      uint64 // total number of sectors written successfully.
	MsecWrite         uint64 // Total number of ms spent by all writes.
	IosInProgress     uint64 // Number of actual I/O requests currently in flight.
	MsecTotal         uint64 // Amount of time during which ios_in_progress >= 1.
	MsecWeightedTotal uint64 // Measure of recent I/O completion time and backlog.
}

func getDiskStats() (*[]DiskStats, error) {
	procDiskStats := "/proc/diskstats"
	if _, err := os.Stat(procDiskStats); err != nil {
		return nil, fmt.Errorf("%s not exists", procDiskStats)
	}

	contents, err := ioutil.ReadFile(procDiskStats)
	if err != nil {
		return nil, err
	}

	var ret []DiskStats

	reader := bufio.NewReader(bytes.NewBuffer(contents))
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		fields := strings.Fields(string(line))
		// shortcut the deduper and just skip disks that
		// haven't done a single read.  This elimiates a bunch
		// of loopback, ramdisk, and cdrom devices but still
		// lets us report on the rare case that we actually use
		// a ramdisk.
		if fields[3] == "0" {
			continue
		}

		size := len(fields)
		// kernel version too low
		if size != 14 {
			continue
		}

		item := DiskStats{}
		for i := 0; i < size; i++ {
			if item.Major, err = strconv.Atoi(fields[0]); err != nil {
				return nil, err
			}

			if item.Minor, err = strconv.Atoi(fields[1]); err != nil {
				return nil, err
			}

			item.Device = fields[2]

			if item.ReadRequests, err = strconv.ParseUint(fields[3], 10, 64); err != nil {
				return nil, err
			}

			if item.ReadMerged, err = strconv.ParseUint(fields[4], 10, 64); err != nil {
				return nil, err
			}

			if item.ReadSectors, err = strconv.ParseUint(fields[5], 10, 64); err != nil {
				return nil, err
			}

			if item.MsecRead, err = strconv.ParseUint(fields[6], 10, 64); err != nil {
				return nil, err
			}

			if item.WriteRequests, err = strconv.ParseUint(fields[7], 10, 64); err != nil {
				return nil, err
			}

			if item.WriteMerged, err = strconv.ParseUint(fields[8], 10, 64); err != nil {
				return nil, err
			}

			if item.WriteSectors, err = strconv.ParseUint(fields[9], 10, 64); err != nil {
				return nil, err
			}

			if item.MsecWrite, err = strconv.ParseUint(fields[10], 10, 64); err != nil {
				return nil, err
			}

			if item.IosInProgress, err = strconv.ParseUint(fields[11], 10, 64); err != nil {
				return nil, err
			}

			if item.MsecTotal, err = strconv.ParseUint(fields[12], 10, 64); err != nil {
				return nil, err
			}

			if item.MsecWeightedTotal, err = strconv.ParseUint(fields[13], 10, 64); err != nil {
				return nil, err
			}

		}
		ret = append(ret, item)
	}
	return &ret, nil
}

type DiskCollector struct {
	host     string
	Interval int64
}

func (dc *DiskCollector) Config(config map[string]interface{}) {
	fmt.Printf("%s config %#v\n", dc.Name(), dc)

	dc.host, _ = os.Hostname()

	if err := mapstructure.Decode(config, dc); err != nil {
		panic(err)
	}

	fmt.Printf("%s config %#v\n", dc.Name(), dc)
}

func (dc *DiskCollector) Detect() bool {
	if runtime.GOOS == "linux" {
		return true
	} else {
		return false
	}
}

func (dc *DiskCollector) Run(c chan *Metric) {

	if !dc.Detect() {
		return
	}

	pdiskstats, _ := getDiskStats()

	for {
		select {
		case <-time.After(time.Second * time.Duration(dc.Interval)):

			diskstats, _ := getDiskStats()

			for k, v := range *diskstats {
				matched, err := regexp.MatchString("^([a-z]+)$", v.Device)

				if err != nil || !matched {
					continue
				}

				c <- Gauge(dc.host + ".disk." + v.Device + ".writes").Record((v.WriteSectors - (*pdiskstats)[k].WriteSectors) / uint64(dc.Interval))
				c <- Gauge(dc.host + ".disk." + v.Device + ".reads").Record((v.ReadSectors - (*pdiskstats)[k].ReadSectors) / uint64(dc.Interval))
				c <- Gauge(dc.host + ".disk." + v.Device + ".iops").Record(v.IosInProgress)

				if v.ReadRequests-(*pdiskstats)[k].ReadRequests != 0 {
					c <- Gauge(dc.host + ".disk." + v.Device + ".latency.read").Record((v.MsecRead - (*pdiskstats)[k].MsecRead) / (v.ReadRequests - (*pdiskstats)[k].ReadRequests))
				} else {
					c <- Gauge(dc.host + ".disk." + v.Device + ".latency.read").Record(0)
				}

				if v.WriteRequests-(*pdiskstats)[k].WriteRequests != 0 {
					c <- Gauge(dc.host + ".disk." + v.Device + ".latency.write").Record((v.MsecWrite - (*pdiskstats)[k].MsecWrite) / (v.WriteRequests - (*pdiskstats)[k].WriteRequests))
				} else {
					c <- Gauge(dc.host + ".disk." + v.Device + ".latency.write").Record(0)
				}
			}
			pdiskstats = diskstats
		}
	}
}

func (*DiskCollector) Name() string {
	return "linux.disk.stats"
}
