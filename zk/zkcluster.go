package zk

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// ZkCluster is a kafka cluster that has a chroot path in Zookeeper.
type ZkCluster struct {
	zone *ZkZone
	name string // cluster name
	path string // cluster's kafka chroot path in zk cluster
}

func (this *ZkCluster) GetTopics(cluster string) []string {
	r := make([]string, 0)
	for name, _ := range this.zone.getChildrenWithData(clusterRoot + BrokerTopicsPath) {
		r = append(r, name)
	}
	return r
}

// returns {brokerId: broker}
func (this *ZkCluster) Brokers() map[string]*Broker {
	r := make(map[string]*Broker)
	for brokerId, brokerInfo := range this.zone.getChildrenWithData(this.path + BrokerIdsPath) {
		broker := newBroker(brokerId)
		broker.from(brokerInfo)

		r[brokerId] = broker
	}

	return r
}

func (this *ZkCluster) Isr(topic string, partitionId int32) []int {
	partitionStateData, _ := this.zone.getData(fmt.Sprintf("%s%s/%s/partitions/%d/state", this.path, BrokerTopicsPath, topic,
		partitionId))
	partitionState := make(map[string]interface{})
	json.Unmarshal(partitionStateData, &partitionState)
	isr := partitionState["isr"].([]interface{})
	r := make([]int, 0, len(isr))
	for _, id := range isr {
		r = append(r, int(id.(float64)))
	}
	sort.Ints(r)

	return r
}

func (this *ZkCluster) Broker(id int) (b *Broker) {
	idStr := strconv.Itoa(id)
	zkData, _ := this.zone.getData(this.path + BrokerIdsPath +
		zkPathSeperator + idStr)
	b = newBroker(idStr)
	b.from(zkData)
	return
}