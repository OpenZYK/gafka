package kafka

import (
	"github.com/funkygao/golib/set"
	log "github.com/funkygao/log4go"
	pool "github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type pubPool struct {
	store *pubStore

	cluster    string
	size       int
	nextId     uint64
	brokerList []string

	syncPool  *pool.ResourcePool
	asyncPool *pool.ResourcePool
}

func newPubPool(store *pubStore, cluster string, brokerList []string, size int) *pubPool {
	this := &pubPool{
		store:      store,
		cluster:    cluster,
		size:       size,
		brokerList: brokerList,
	}
	this.buildPools()

	return this
}

func (this *pubPool) buildPools() {
	// idleTimeout=0 means each kafka conn will last forever
	this.syncPool = pool.NewResourcePool(this.syncProducerFactory,
		this.size, this.size, 0)
	this.asyncPool = pool.NewResourcePool(this.asyncProducerFactory,
		this.size, this.size, 0)
}

func (this *pubPool) RefreshBrokerList(brokerList []string) {
	if len(brokerList) == 0 {
		if len(this.brokerList) > 0 {
			log.Warn("%s meta store found empty broker list, refresh refused", this.cluster)
		}
		return
	}

	setOld, setNew := set.NewSet(), set.NewSet()
	for _, b := range this.brokerList {
		setOld.Add(b)
	}
	for _, b := range brokerList {
		setNew.Add(b)
	}

	if !setOld.Equal(setNew) {
		log.Info("%s broker list from %+v to %+v", this.cluster, this.brokerList, brokerList)

		// rebuild the kafka conn pool
		this.brokerList = brokerList
		this.Close()
		this.buildPools()
	}
}

func (this *pubPool) Close() {
	this.syncPool.Close()
	this.syncPool = nil

	this.asyncPool.Close()
	this.asyncPool = nil
}

func (this *pubPool) GetSyncProducer() (*syncProducerClient, error) {
	ctx := context.Background()
	k, err := this.syncPool.Get(ctx)
	if err != nil {
		return nil, err
	}

	return k.(*syncProducerClient), nil
}

func (this *pubPool) GetAsyncProducer() (*asyncProducerClient, error) {
	ctx := context.Background()
	k, err := this.asyncPool.Get(ctx)
	if err != nil {
		return nil, err
	}

	return k.(*asyncProducerClient), nil
}
