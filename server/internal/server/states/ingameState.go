package states

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/guruorgoru/go-mmo/server/internal/objects"
	"github.com/guruorgoru/go-mmo/server/internal/server"
	"github.com/guruorgoru/go-mmo/server/pkg/packets"
)

type InGame struct {
	client                 server.ClientInterfacer
	player                 *objects.Player
	logger                 *log.Logger
	cancelPlayerUpdateLoop context.CancelFunc
}

func (ig *InGame) Name() string {
	return "InGame"
}

func (ig *InGame) SetClient(client server.ClientInterfacer) {
	ig.client = client
	logPrefix := fmt.Sprintf("[%v] Client %v: ", ig.Name(), ig.client.GetId())
	ig.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
}

func (ig *InGame) OnEntry() {
	ig.logger.Printf("Adding %v to the sharedGameObject", ig.player.Name)
	ig.player.X = rand.Float64() * 1000
	ig.player.Y = rand.Float64() * 1000
	ig.player.Radius = 20.0
	ig.player.Speed = 150.0

	go ig.client.SharedGameObjects().Players.Add(ig.player, ig.client.GetId())
	ig.client.Send(packets.NewPlayer(ig.client.GetId(), ig.player))
}

func (ig *InGame) Handle(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_Player:
		ig.handlePlayer(senderId, msg)
	case *packets.Packet_PlayerDirection:
		ig.handlePlayerDirection(senderId, msg)
	case *packets.Packet_ChatMsg:
		ig.handleChatMessage(senderId, msg)
	}
}

func (ig *InGame) handlePlayerDirection(senderId uint64, msg *packets.Packet_PlayerDirection) {
	if senderId == ig.client.GetId() {
		ig.player.Direction = msg.PlayerDirection.Direction
		
		if ig.cancelPlayerUpdateLoop == nil {
			ctx, cancel := context.WithCancel(context.Background())
			ig.cancelPlayerUpdateLoop = cancel
			go ig.updatePlayerLoop(ctx)
		}
	}
}

func (ig *InGame) handleChatMessage(senderId uint64, msg packets.Msg) {
	if senderId == ig.client.GetId() {
		ig.client.Broadcast(msg)
	} else {
		ig.client.SendAs(msg, senderId)
	}
}

func (ig *InGame) handlePlayer(senderId uint64, message *packets.Packet_Player) {
	if senderId == ig.client.GetId() {
		ig.logger.Println("Received player message from our own client, ignoring")
		return
	}
	ig.client.SendAs(message, senderId)
}

func (ig *InGame) updatePlayerLoop(ctx context.Context) {
	const delta float64 = 0.05
	ticker := time.NewTicker(time.Duration(delta*1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ig.syncPlayer(delta)
		case <-ctx.Done():
			return
		}
	}
}

func (ig *InGame) OnExit() {
	if ig.cancelPlayerUpdateLoop != nil {
		ig.cancelPlayerUpdateLoop()
	}
	ig.client.SharedGameObjects().Players.Delete(ig.client.GetId())
}

func (ig *InGame) syncPlayer(delta float64) {
	newX := ig.player.X + ig.player.Speed*math.Cos(ig.player.Direction)*delta
	newY := ig.player.Y + ig.player.Speed*math.Sin(ig.player.Direction)*delta

	ig.player.X = newX
	ig.player.Y = newY

	updatePacket := packets.NewPlayer(ig.client.GetId(), ig.player)
	ig.client.Broadcast(updatePacket)
	go ig.client.Send(updatePacket)
}
