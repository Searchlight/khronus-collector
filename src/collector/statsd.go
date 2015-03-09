package collector

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"

	. "github.com/despegar/khronus-go-client"
	"github.com/mitchellh/mapstructure"
)

type StatsdCollector struct {
	Interval      int64
	Port          float64
	CounterPrefix string
	GaugesPrefix  string
	TimersPrefix  string
}

func (stdc *StatsdCollector) Config(config map[string]interface{}) {
	fmt.Printf("%s config %#v\n", stdc.Name(), stdc)

	if err := mapstructure.Decode(config, stdc); err != nil {
		panic(err)
	}

	if stdc.Port == 0 {
		stdc.Port = 8125
	}

	fmt.Printf("%s config %#v\n", stdc.Name(), stdc)
}

func (stdc *StatsdCollector) Detect() bool {
	return true
}

func handleMessage(buf *bytes.Buffer, c chan *Metric) {
	var sanitizeRegexp = regexp.MustCompile("[^a-zA-Z0-9\\-_\\.:\\|@]")
	var packetRegexp = regexp.MustCompile("([a-zA-Z0-9\\-_\\.]+)((:[0-9]+\\|(c|g|ms)(\\|@([0-9\\.]+))?)+)")
	s := sanitizeRegexp.ReplaceAllString(buf.String(), "")

	for _, item := range packetRegexp.FindAllStringSubmatch(s, -1) {
		values := []uint64{}

		for _, v := range strings.Split(item[2], ":")[1:] {
			sp := strings.Split(v, "|")
			value, _ := strconv.ParseUint(sp[0], 10, 64)

			if len(sp) == 3 {
				s, _ := strconv.ParseFloat(sp[2][1:], 64)
				num := int(math.Mod(float64(value), 1/float64(s)))
				value := uint(float64(value) * s)

				for i := 0; i < num; i++ {
					values = append(values, uint64(value))
				}
			} else {
				values = append(values, uint64(value))
			}
		}

		switch item[4] {
		case "c":
			c <- Counter(item[1]).Record(values...)
		case "g":
			c <- Gauge(item[1]).Record(values...)
		case "ms":
			c <- Timer(item[1]).Record(values...)
		}

		/*
			fmt.Println(
				fmt.Sprintf("Packet: bucket = %s, value = %+v, modifier = %s, sampling = %f\n",
					item[1], values, item[4], sampleRate))
		*/
	}
}

func (stdc *StatsdCollector) Run(c chan *Metric) {

	if !stdc.Detect() {
		return
	}

	address, err := net.ResolveUDPAddr("udp", ":8125")

	if err != nil {
		fmt.Println(err)
		return
	}

	listener, err := net.ListenUDP("udp", address)
	defer listener.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		message := make([]byte, 1024)
		n, _, error := listener.ReadFrom(message)
		if error != nil {
			continue
		}
		buf := bytes.NewBuffer(message[0:n])
		go handleMessage(buf, c)
	}
}

func (*StatsdCollector) Name() string {
	return "statsd"
}
