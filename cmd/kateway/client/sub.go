package main

import (
	"flag"
	"fmt"

	"github.com/funkygao/gafka/cmd/kateway/api"
)

var (
	addr  string
	n     int
	appid string
	group string
)

func init() {
	flag.StringVar(&addr, "addr", "http://10.1.114.159:9192", "sub kateway addr")
	flag.StringVar(&group, "g", "mygroup1", "consumer group name")
	flag.StringVar(&appid, "appid", "", "consume whose topic")
	flag.IntVar(&n, "n", 20, "run sub how many times")
	flag.Parse()
}

func main() {
	c := api.NewClient("demo", nil)
	c.Connect(addr)
	i := 0
	err := c.Subscribe(appid, "foobar", "ver", group, func(statusCode int, msg []byte) error {
		fmt.Printf("%10d: %s\n", i+1, string(msg))

		i++
		if i >= n {
			return api.ErrSubStop
		}

		return nil
	})

	fmt.Println(err)
}
