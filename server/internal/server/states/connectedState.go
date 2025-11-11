package states

import (
	"fmt"
	"log"

	"github.com/guruorgoru/go-mmo/server/internal/server"
	"github.com/guruorgoru/go-mmo/server/pkg/packets"
)

type Connected struct {
	client server.ClientInterfacer
	logger *log.Logger
}

func (c *Connected) Name() string {
	return "Connected"
}

// Inject client and set everything up
func (c *Connected) SetClient(newClient server.ClientInterfacer) {
	c.client = newClient
	logPrefix := fmt.Sprintf("[%v] Client %v: ", c.Name(), c.client.GetId())
	c.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
}

func (c *Connected) OnEntry() {
	// Newly connected client will get his/her Id first
	c.client.Send(packets.NewId(c.client.GetId()))
}

func (c *Connected) OnExit() {
	// TODO
}

func (c *Connected) Handle(senderId uint64, msg packets.Msg) {
	if senderId == c.client.GetId() {
		c.client.Broadcast(msg)
	} else {
		c.client.SendAs(msg, senderId)
	}
}
