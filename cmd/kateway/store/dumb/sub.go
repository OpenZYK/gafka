package dumb

import (
	"sync"

	"github.com/Shopify/sarama"
	"github.com/funkygao/gafka/cmd/kateway/store"
)

type subStore struct {
	fetcher *consumerFetcher
}

func NewSubStore(wg *sync.WaitGroup, shutdownCh <-chan struct{},
	closedConnCh <-chan string, debug bool) *subStore {
	return &subStore{
		fetcher: &consumerFetcher{
			ch: make(chan *sarama.ConsumerMessage, 1000),
		},
	}
}

func (this *subStore) Start() (err error) {
	msg := &sarama.ConsumerMessage{
		Topic: "hello",
		Key:   []byte("world"),
		Value: []byte("hello from dumb fetcher"),
	}

	go func() {
		for {
			this.fetcher.ch <- msg
		}
	}()

	return
}

func (this *subStore) Stop() {}

func (this *subStore) Name() string {
	return "dumb"
}

func (this *subStore) KillClient(remoteAddr string) {

}
func (this *subStore) Fetch(cluster, topic, group, remoteAddr, reset string) (store.Fetcher, error) {
	return this.fetcher, nil
}
