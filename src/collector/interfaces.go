package collector

import . "github.com/Searchlight/khronus-go-client"

type Collector interface {
	Run(c chan *Metric)
	Detect() bool
	Config(map[string]interface{})
	Name() string
}

type Output interface {
	Config(config map[string]interface{})
	Run(cd chan *Metric)
}
