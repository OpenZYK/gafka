package main

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/funkygao/log4go"
	"github.com/wvanbergen/kafka/consumergroup"
)

type subPool struct {
	gw *Gateway

	// {topic: {group: {client: consumerGroup}}}
	cgs     map[string]map[string]map[string]*consumergroup.ConsumerGroup
	cgsLock sync.RWMutex
}

func newSubPool(gw *Gateway) *subPool {
	return &subPool{
		gw:  gw,
		cgs: make(map[string]map[string]map[string]*consumergroup.ConsumerGroup),
	}
}

func (this *subPool) PickConsumerGroup(topic, group,
	client string) (cg *consumergroup.ConsumerGroup, err error) {
	this.cgsLock.Lock()
	defer this.cgsLock.Unlock()

	var present bool
	if _, present = this.cgs[topic]; !present {
		this.cgs[topic] = make(map[string]map[string]*consumergroup.ConsumerGroup)
	}
	if _, present = this.cgs[topic][group]; !present {
		this.cgs[topic][group] = make(map[string]*consumergroup.ConsumerGroup)
	}
	cg, present = this.cgs[topic][group][client]
	if present {
		return
	}

	if this.gw.metaStore.OnlineConsumersCount(topic, group) >= len(this.gw.metaStore.Partitions(topic)) {
		err = ErrTooManyConsumers

		return
	}

	// create the consumer group for this client
	cf := consumergroup.NewConfig()
	cf.ChannelBufferSize = 0
	cf.Offsets.Initial = sarama.OffsetOldest
	cf.Offsets.CommitInterval = options.offsetCommitInterval
	// time to wait for all the offsets for a partition to be processed after stopping to consume from it.
	cf.Offsets.ProcessingTimeout = time.Second * 10
	cf.Zookeeper.Chroot = this.gw.metaStore.ZkChroot()
	for i := 0; i < 3; i++ {
		cg, err = consumergroup.JoinConsumerGroup(group, []string{topic},
			this.gw.metaStore.ZkAddrs(), cf)
		if err == nil {
			this.cgs[topic][group][client] = cg
			break
		}
	}

	return
}

func (this *subPool) KillClient(topic, group, client string) {
	this.cgs[topic][group][client].Close() // will flush offset
	delete(this.cgs[topic][group], client)
	log.Info("consumer %s{topic:%s, group:%s} closed", client, topic, group)
}

func (this *subPool) Start() {
	for {
		select {
		case <-this.gw.shutdownCh:
			log.Info("sub pool shutdown")
			this.Stop()
			this.gw.wg.Done()
			break

		}
	}

}

func (this *subPool) Stop() {
	var wg sync.WaitGroup
	for topic, ts := range this.cgs {
		for group, gs := range ts {
			for client, c := range gs {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := c.Close(); err != nil {
						// will commit the offset
						log.Error("{topic:%s, group:%s, client:%s}: %v", topic,
							group, client, err)
					}
				}()
			}
		}
	}

	// wait for all consumers commit offset
	wg.Wait()

	// reinit the vars
	this.cgs = make(map[string]map[string]map[string]*consumergroup.ConsumerGroup)
}