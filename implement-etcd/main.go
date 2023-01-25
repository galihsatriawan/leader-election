package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func main() {
	var name = flag.String("name", "tester", "give a name")
	flag.Parse()
	// Create a etcd client
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// create a sessions to elect a Leader
	s, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	keyPrefix := "/sharding-service/"
	e := concurrency.NewElection(s, keyPrefix)
	ctx := context.Background()

	var resp *clientv3.GetResponse
	go func() {
		temp := <-e.Observe(ctx)
		resp = &temp
		fmt.Println(resp)
	}()
	// Elect a leader (or wait that the leader resign)
	if err := e.Campaign(ctx, "2"); err != nil {
		log.Fatal(err)
	}
	if resp != nil {
		e.Proclaim(ctx, string(resp.Kvs[0].Value))
	}
	leader, err := e.Leader(ctx)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(leader.Kvs[0].Value))
	e.Proclaim(ctx, "4")
	fmt.Println("leader election for ", *name)
	fmt.Println("Do some work in", *name)
	time.Sleep(5 * time.Second)
	leader, err = e.Leader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(leader.Kvs[0].Value))
	if err := e.Resign(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("resign ", *name)
}
