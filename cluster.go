package rediscluster

import (
	"github.com/fzzbt/radix/redis"
	"github.com/joaojeronimo/go-crc16"
	"strconv"
	"strings"
)

func clusterSlot(key string) int {
	return (int(crc16.Crc16(key)) % 4096)
}

type SlotInterval struct {
	LowerLimit int
	UpperLimit int
}

func ParseNode(unparsedNode string) (node Node) {
	parts := strings.Split(unparsedNode, " ")
	node.Name = parts[0]
	node.Address = parts[1]
	node.Flags = parts[2]
	//node.LastPingSent, _ = strconv.Atoi(parts[4])
	//node.LastPongReceived, _ = strconv.Atoi(parts[5])
	node.State = parts[6]

	slots := strings.Split(parts[7], "-")
	var slotInterval SlotInterval
	slotInterval.LowerLimit, _ = strconv.Atoi(slots[0])
	slotInterval.UpperLimit, _ = strconv.Atoi(slots[1])
	node.Slots = slotInterval
	return
}

type Node struct {
	Name             string
	Address          string
	Client           *redis.Client
	Flags            string
	LastPingSent     int
	LastPongReceived int
	State            string
	Slots            SlotInterval
}

type Cluster struct {
	nodes []Node
}

func (n *Node) Connect() {
	conf := redis.DefaultConfig()
	conf.Address = n.Address
	n.Client = redis.NewClient(conf)
}

func discoverNodes(firstLink string) (cluster Cluster) {
	conf := redis.DefaultConfig()
	conf.Address = firstLink
	c := redis.NewClient(conf)
	s, err := c.Cluster("nodes").Str()
	if err != nil {
		return
	}
	unparsedNodes := strings.Split(s, "\n")
	var parsedNodes []Node
	for i := 0; i < len(unparsedNodes)-1; i++ { // -1 because the last line is empty
		parsedNode := ParseNode(unparsedNodes[i])
		if parsedNode.Address == ":0" {
			parsedNode.Address = firstLink
			parsedNode.Client = c
		} else {
			parsedNode.Connect()
		}
		parsedNodes = append(parsedNodes, parsedNode)
	}
	cluster.nodes = parsedNodes
	return
}

func NewCluster(firstLink string) (cc Cluster) {
	return discoverNodes(firstLink)
}

func (cc *Cluster) Call(command string, args ...interface{}) (reply *redis.Reply) {
	slot := clusterSlot(args[0].(string))
	var slots SlotInterval
	for i := 0; i < len(cc.nodes); i++ {
		slots = cc.nodes[i].Slots
		if slot > slots.LowerLimit && slot < slots.UpperLimit {
			reply = cc.nodes[i].Client.Call(command, args...)
			return
		}
	}
	return
}

func (cc *Cluster) AsyncCall(command string, args ...interface{}) (future redis.Future) {
	slot := clusterSlot(args[0].(string))
	var slots SlotInterval
	for i := 0; i < len(cc.nodes); i++ {
		slots = cc.nodes[i].Slots
		if slot >= slots.LowerLimit && slot =< slots.UpperLimit {
			future = cc.nodes[i].Client.AsyncCall(command, args...)
			return
		}
	}
	return
}

func (cc *Cluster) Close() bool {
	for i := 0; i < len(cc.nodes); i++ {
		cc.nodes[i].Client.Close()
	}
	return true
}
