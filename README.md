# Go Redis Cluster

A [Redis](http://redis.io/) [Cluster](http://redis.io/topics/cluster-spec) Client for the Go Programming Language.

(Well, not really. It's only a thin wrapper above the amazing [Radix](https://github.com/fzzbt/radix/) module. It even uses the Radix module, borrows heavily from it, and the Radix guys did 99% of the work. This is only a thin wrapper that discovers your cluster, connects to all the nodes, and does client-side sharding.)

## Installing it

Since this is a wrapper above Radix, you'll need to install it:
```
go get github.com/fzzbt/radix/redis
```
And then you can install Go Redis Cluster:
```
go get github.com/joaojeronimo/go-redis-cluster
```

## Using it

You'll need the link of one member of the cluster, and the rest will be automatically discovered, and everything will be dealt with (slots, etc). After that, you can use it pretty much like you use the [Radix](https://github.com/fzzbt/radix/) module, except you cannot use the obvious commands that (a) [affect multiple keys](http://redis.io/topics/transactions) (b) [Publish/Subscribe](http://redis.io/topics/pubsub).

Redis Cluster only implements a subset of the commands: "all the **single keys** commands available in the non distributed version of Redis" (read more about this in the **Implemented subset** section of the [Cluster Spec](http://redis.io/topics/cluster-spec))

Other than that, it's almost exactly the same as using the Radix module, even the commands' names are the same:

```go
package main

import (
	"github.com/joaojeronimo/go-redis-cluster"
)

func main() {
	cc := rediscluster.NewCluster("localhost:6379") // Connect to the whole cluster only knowing the link of one member
	defer cc.Close()                                // Close it when main returns

	// the usual commands:
	cc.Hset("someHash", "someKey", "someValue")
	theValue := cc.Hget("someHash", "someKey").String()
	if theValue == "someValue" {
		println("OK")
	}

	// Async stuff too:
	cc.Sadd("someSet", "value1", "value2", "value3")
	future := cc.AsyncScard("someSet")
	card, err := (<-future).Int()
	if err != nil {
		panic(err)
	}
	if 3 == card {
		println("OK")
	}
}
```

## Other stuff

I also created a [similar module for Node.JS](https://github.com/joaojeronimo/node_redis_cluster), we use both of them in production, as well as Redis 3.0 that is unstable but passes the tests. So far so good :)