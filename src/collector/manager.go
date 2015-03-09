package collector

import (
	"fmt"
	"os"
	"os/signal"

	. "github.com/despegar/khronus-go-client"
)

type Manager struct {
	outchan chan *Metric
	colchan chan *Metric
	outputs map[string]interface {
		Output
	}
	collectors map[string]interface {
		Collector
	}
}

func (m *Manager) Config(mc map[string]interface{}) {

	m.outchan = make(chan *Metric, 100)
	m.colchan = make(chan *Metric, 100)

	fmt.Printf("Configuring khronus collector manager\n")

	m.collectors = map[string]interface {
		Collector
	}{
		"CpuCollector":    &CpuCollector{},
		"MemCollector":    &MemCollector{},
		"DiskCollector":   &DiskCollector{},
		"NetCollector":    &NetCollector{},
		"StatsdCollector": &StatsdCollector{},
	}

	m.outputs = map[string]interface {
		Output
	}{
		"KhronusOutput": &KhronusOutput{},
	}

	fmt.Printf("Configuring Outputs %#v\n", mc)

	for on, oc := range mc["outputs"].(map[string]interface{}) {
		m.outputs[on].(interface {
			Output
		}).Config(map[string]interface{}(oc.(map[string]interface{})))
	}

	fmt.Printf("Configuring Collectors %#v\n", mc)

	for cn, cc := range mc["collectors"].(map[string]interface{}) {
		m.collectors[cn].(interface {
			Collector
		}).Config(map[string]interface{}(cc.(map[string]interface{})))
	}

	fmt.Printf("Config Settings %#v\n", mc)
}

func (m *Manager) Run() {
	fmt.Println("Running khronus collector manager")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)

	for _, o := range m.outputs {
		go interface {
			Output
		}(o).Run(m.outchan)
	}

	for _, c := range m.collectors {
		go interface {
			Collector
		}(c).Run(m.colchan)
	}

	done := false

	for {
		if done == true {
			os.Exit(0)
		}
		select {
		case mdp := <-m.colchan:
			{
				m.outchan <- mdp
			}
		case signal := <-ch:
			{
				fmt.Printf("Khronus manager ends by signal: %s\n", signal)
				done = true
				break
			}
		}
	}
}
