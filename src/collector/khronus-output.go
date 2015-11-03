package collector

import (
	. "github.com/Searchlight/khronus-go-client"
	"github.com/mitchellh/mapstructure"
)

type KhronusOutput struct {
	Urls     []string
	Prefix   string
	Interval uint64
	k        Client
}

func (ko *KhronusOutput) Config(config map[string]interface{}) {
	if err := mapstructure.Decode(config, ko); err != nil {
		panic(err)
	}

	ko.k = Client{}

	ko.k.Config().Interval(ko.Interval)

	for k, v := range ko.Urls {
		ko.Urls[k] = v + ko.Prefix
	}

	ko.k.Config().Urls(ko.Urls)

}

func (ko *KhronusOutput) Run(cd chan *Metric) {
	ko.k.Config().Channel(cd)
}

func (ko *KhronusOutput) Name() string {
	return `output.khronus`
}
