package beater

import (
	"fmt"
	"time"
        "os"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/toravir/cborbeat/config"
        csd "github.com/toravir/csd/libs"
)

// cborbeat configuration.
type cborbeat struct {
        inputFile string
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of cborbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

        info, err := os.Stat(c.File)
        if err == nil && info.IsDir() {
		return nil, fmt.Errorf("Error input file is a DIR !! %s pls check cborbeat.yml. %v", c.File, err)
        }

	bt := &cborbeat{
                inputFile: c.File,
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts cborbeat.
func (bt *cborbeat) Run(b *beat.Beat) error {
        logp.Info("cborbeat is running! Hit CTRL-C to stop it. inputFile:%v ...", bt.inputFile)

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

        rdr, err := NewFollowReader(bt.inputFile, true, bt.done)
        var decoder *csd.Decoder = nil
        if err == nil {
            decoder = csd.NewDecoder(rdr)
        }

	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
                if decoder == nil {
                    rdr, err := NewFollowReader(bt.inputFile, true, bt.done)
                    if err == nil {
                        decoder = csd.NewDecoder(rdr)
                    }
                    if decoder == nil {
                        //Try again next time
                        continue
                    }
                }
                ll, _ := decoder.SafeNext()
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
                                "log": ll,
			},
		}
		bt.client.Publish(event)
		counter++
	}
}

// Stop stops cborbeat.
func (bt *cborbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
