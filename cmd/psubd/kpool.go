package main

import (
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/golib/pool"
	"github.com/funkygao/golib/set"
	log "github.com/funkygao/log4go"
)

// kafka client
type kclient struct {
	id   uint64
	pool *kpool

	sarama.Client
}

func (this *kclient) Close() {
	this.Client.Close()
	//this.pool.pool.Put(nil)
}

func (this *kclient) Id() uint64 {
	return this.id
}

func (this *kclient) IsOpen() bool {
	return !this.Client.Closed()
}

func (this *kclient) Recycle() {
	if this.Client.Closed() {
		this.pool.pool.Kill(this)
		this.pool.pool.Put(nil)
	} else {
		this.pool.pool.Put(this)
	}

}

// kafka client pool
type kpool struct {
	brokerList []string
	pool       *pool.ResourcePool
	nextId     uint64
}

func newKpool(brokerList []string) *kpool {
	this := &kpool{
		brokerList: brokerList,
	}
	factory := func() (pool.Resource, error) {
		conn := &kclient{
			pool: this,
			id:   atomic.AddUint64(&this.nextId, 1),
		}

		var err error
		t1 := time.Now()
		conn.Client, err = sarama.NewClient(brokerList, sarama.NewConfig())
		if err == nil {
			log.Debug("kafka connected[%d]: %+v %s", conn.id, brokerList, time.Since(t1))
		} else {
			log.Error("kafka %+v: %v %s", brokerList, err, time.Since(t1))
		}

		return conn, err
	}

	this.pool = pool.NewResourcePool("kafka", factory,
		50, 50, 0, time.Second*10, time.Second*5)

	return this
}

func (this *kpool) Close() {
	this.pool.Close()
}

func (this *kpool) Get() (*kclient, error) {
	k, err := this.pool.Get()
	if err != nil {
		return nil, err
	}

	return k.(*kclient), nil
}

func (this *kpool) RefreshBrokerList(brokerList []string) {
	setOld, setNew := set.NewSet(), set.NewSet()
	for _, b := range this.brokerList {
		setOld.Add(b)
	}
	for _, b := range brokerList {
		setNew.Add(b)
	}
	if !setOld.Equal(setNew) {
		log.Warn("brokers change: %+v -> %+v", this.brokerList, brokerList)
		// TODO
	}

}