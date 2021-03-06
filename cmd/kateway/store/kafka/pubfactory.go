package kafka

import (
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/gafka/cmd/kateway/store"
	log "github.com/funkygao/log4go"
	pool "github.com/youtube/vitess/go/pools"
)

func (this *pubPool) syncProducerFactory() (pool.Resource, error) {
	if len(this.brokerList) == 0 {
		return nil, store.ErrEmptyBrokers
	}

	spc := &syncProducerClient{
		pool: this,
		id:   atomic.AddUint64(&this.nextId, 1),
	}

	var err error
	t1 := time.Now()
	cf := sarama.NewConfig()
	cf.Net.DialTimeout = time.Second * 10
	cf.Net.ReadTimeout = time.Second * 10
	cf.Net.WriteTimeout = time.Second * 10

	cf.Metadata.RefreshFrequency = time.Minute * 10
	cf.Metadata.Retry.Max = 3
	cf.Metadata.Retry.Backoff = time.Second

	cf.Producer.RequiredAcks = sarama.WaitForLocal
	cf.Producer.Partitioner = sarama.NewHashPartitioner
	cf.Producer.Return.Successes = false
	cf.Producer.Retry.Max = 3
	//cf.Producer.Compression = sarama.CompressionSnappy

	cf.ClientID = this.store.hostname

	cf.ChannelBufferSize = 256

	spc.SyncProducer, err = sarama.NewSyncProducer(this.brokerList, cf)
	if err != nil {
		return nil, err
	}

	log.Trace("cluster[%s] kafka connected[%d]: %+v %s",
		this.cluster, spc.id, this.brokerList, time.Since(t1))

	return spc, err
}

func (this *pubPool) asyncProducerFactory() (pool.Resource, error) {
	if len(this.brokerList) == 0 {
		return nil, store.ErrEmptyBrokers
	}

	apc := &asyncProducerClient{
		pool: this,
		id:   atomic.AddUint64(&this.nextId, 1),
	}

	var err error
	t1 := time.Now()
	cf := sarama.NewConfig()
	cf.Metadata.RefreshFrequency = time.Minute // TODO
	cf.Metadata.Retry.Max = 3                  //

	cf.Producer.Flush.Frequency = time.Second * 10
	cf.Producer.Flush.Messages = 1000
	cf.Producer.Flush.MaxMessages = 0 // unlimited

	cf.Producer.RequiredAcks = sarama.NoResponse
	cf.Producer.Partitioner = sarama.NewHashPartitioner
	cf.Producer.Retry.Max = 3
	//cf.Producer.Compression = sarama.CompressionSnappy TODO

	cf.ClientID = this.store.hostname

	apc.AsyncProducer, err = sarama.NewAsyncProducer(this.brokerList, cf)
	if err != nil {
		return nil, err
	}

	log.Trace("cluster[%s] kafka connected[%d]: %+v %s",
		this.cluster, apc.id, this.brokerList, time.Since(t1))

	// TODO
	go func() {
		// messages will only be returned here after all retry attempts are exhausted.
		for err := range apc.Errors() {
			log.Error("cluster[%s] async producer: %v", this.cluster, err)
		}
	}()

	return apc, err
}
