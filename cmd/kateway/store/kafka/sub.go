package kafka

import (
	l "log"
	"os"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/funkygao/gafka/cmd/kateway/store"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/golib/color"
	log "github.com/funkygao/log4go"
	"github.com/wvanbergen/kafka/consumergroup"
)

type consumerFetcher struct {
	*consumergroup.ConsumerGroup
}

type subStore struct {
	shutdownCh   <-chan struct{}
	closedConnCh <-chan string
	wg           *sync.WaitGroup
	hostname     string

	subPool *subPool
}

func NewSubStore(wg *sync.WaitGroup, shutdownCh <-chan struct{},
	closedConnCh <-chan string, debug bool) *subStore {
	if debug {
		sarama.Logger = l.New(os.Stdout, color.Blue("[Sarama]"),
			l.LstdFlags|l.Lshortfile)
	}

	return &subStore{
		hostname:     ctx.Hostname(),
		wg:           wg,
		shutdownCh:   shutdownCh,
		closedConnCh: closedConnCh,
	}
}

func (this *subStore) Start() (err error) {
	this.wg.Add(1)
	defer this.wg.Done()

	this.subPool = newSubPool()

	go func() {
		var remoteAddr string
		for {
			select {
			case <-this.shutdownCh:
				this.subPool.Stop()
				log.Trace("kafka sub store stopped")
				return

			case remoteAddr = <-this.closedConnCh:
				log.Trace("%s closed, killing sub", remoteAddr)
				this.subPool.killClient(remoteAddr)

			}
		}
	}()

	return
}

func (this *subStore) Stop() {
	// TODO
}

func (this *subStore) Name() string {
	return "kafka"
}

func (this *subStore) KillClient(remoteAddr string) {
	this.subPool.killClient(remoteAddr)
}

func (this *subStore) Fetch(cluster, topic, group, remoteAddr, reset string) (store.Fetcher, error) {
	cg, err := this.subPool.PickConsumerGroup("", topic, group, remoteAddr, reset)
	if err != nil {
		return nil, err
	}

	return &consumerFetcher{
		ConsumerGroup: cg,
	}, nil
}
