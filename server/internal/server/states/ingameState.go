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
	ig.player.Radius = 20.0
	ig.player.X, ig.player.Y = objects.SpawnCoords(ig.player.Radius, ig.client.SharedGameObjects().Players, nil)
	ig.player.Speed = 150.0

	go ig.client.SharedGameObjects().Players.Add(ig.player, ig.client.GetId())
	ig.client.Send(packets.NewPlayer(ig.client.GetId(), ig.player))

	// Send spores in bacthes in backgrond to the client

	go ig.spawnInitialSpores(20, 40*time.Millisecond)
}

func (ig *InGame) Handle(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_Player:
		ig.handlePlayer(senderId, msg)
	case *packets.Packet_PlayerDirection:
		ig.handlePlayerDirection(senderId, msg)
	case *packets.Packet_ChatMsg:
		ig.handleChatMessage(senderId, msg)
	case *packets.Packet_SporeConsumed:
		ig.handleSporeConsumedMessage(senderId, msg)
	case *packets.Packet_PlayerConsumed:
		ig.handlePlayerConsumedMessage(senderId, msg)
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

func (ig *InGame) spawnInitialSpores(batchSize int, batchDelay time.Duration) {
	sporeBatch := make(map[uint64]*objects.Spore, batchSize)

	ig.client.SharedGameObjects().Spores.ForEach(func(u uint64, s *objects.Spore) {
		sporeBatch[u] = s

		if len(sporeBatch) >= batchSize {
			ig.client.Send(packets.NewSporeBatch(sporeBatch))
			sporeBatch = make(map[uint64]*objects.Spore, batchSize)

			time.Sleep(batchDelay)
		}
	})

	if len(sporeBatch) > 0 {
		ig.client.Send(packets.NewSporeBatch(sporeBatch))
	}
}

// Below will be the helpers for handling the spore consumption

// Check whether the spore client ate exists or not, he can lie
func (ig *InGame) getSpore(sporeId uint64) (*objects.Spore, error) {
	spore, exists := ig.client.SharedGameObjects().Spores.GetObjById(sporeId)
	if !exists {
		return nil, fmt.Errorf("spore with the Id %v doesn't exists, client lied to us :( ", sporeId)
	}
	return spore, nil
}

// Validate whether the player was close the the spore when he ate it
func (ig *InGame) validatePlayerIsCloseToSpore(objX, objY, objRadius, buffer float64) error {
	distanceX := ig.player.X - objX
	distanceY := ig.player.Y - objY
	distanceSquared := distanceX*distanceX + distanceY*distanceY

	tresholdDistance := ig.player.Radius + objRadius + buffer
	tresholdDistanceSquared := tresholdDistance * tresholdDistance
	
	if distanceSquared > tresholdDistanceSquared {
		return fmt.Errorf("player lied again, he is too far from the object. (distance: %f, treshold: %f >> Both in squared formj)", distanceSquared, tresholdDistanceSquared)
	}

	return nil
}

// Helpers for consumption function

func radiusToMass(radius float64) float64 {
	return math.Pi * radius * radius
}

func massToRadius(mass float64) float64 {
	radiusSquared := mass/math.Pi
	return math.Sqrt(radiusSquared)
}

// Actual consumption function

func (ig *InGame) handleSporeConsumedMessage(senderId uint64, msg *packets.Packet_SporeConsumed) {
	if senderId != ig.client.GetId() {
		ig.client.SendAs(msg, senderId)
		return
	}

	genericError := "spore consumption not verified: "

	spore, err := ig.getSpore(msg.SporeConsumed.SporeId)
	if err != nil {
		ig.logger.Println(genericError, err)
		return
	}

	err = ig.validatePlayerIsCloseToSpore(spore.X, spore.Y, spore.Radius, 10)
	if err != nil {
		ig.logger.Println(genericError, err)
		return
	}

	// everything's good, increase the player
	sporeMass := radiusToMass(spore.Radius)
	oldPlayerMass := radiusToMass(ig.player.Radius)
	newPlayerMass := oldPlayerMass + sporeMass

	ig.player.Radius = massToRadius(newPlayerMass)

	go ig.client.SharedGameObjects().Spores.Delete(msg.SporeConsumed.SporeId)
	ig.client.Broadcast(msg)
}

func (ig *InGame) handlePlayerConsumedMessage(senderId uint64, msg *packets.Packet_PlayerConsumed) {
	if senderId != ig.client.GetId() {
		ig.client.SendAs(msg, senderId)

		if msg.PlayerConsumed.PlayerId == ig.client.GetId() {
			ig.logger.Println("You lost try again, respawning you in a random location")
			ig.client.SetState(&InGame{
				player: &objects.Player{
					Name: ig.player.Name,
				},
			})
		}
		return
	}

	genericError := "error consuming player: "

	otherPlayer, err := ig.getOtherPlayer(msg.PlayerConsumed.PlayerId)
	if err != nil {
		log.Println(genericError, err)
		return
	}

	ourMass := radiusToMass(ig.player.Radius)
	otherMass := radiusToMass(otherPlayer.Radius)
	if ourMass <= otherMass*1.2{
		ig.logger.Println(genericError, "you arent big enough to consume this player")
		return
	}

	err = ig.validatePlayerIsCloseToSpore(otherPlayer.X, otherPlayer.Y, otherPlayer.Radius, 10)
	if err != nil {
		ig.logger.Println(genericError, err)
	}

	ourNewMass := ourMass + otherMass
	ig.player.Radius = massToRadius(ourNewMass)

	go ig.client.SharedGameObjects().Players.Delete(msg.PlayerConsumed.PlayerId)
	ig.client.Broadcast(msg)
}

func (ig *InGame) getOtherPlayer(otherId uint64) (*objects.Player, error) {
	other, exists := ig.client.SharedGameObjects().Players.GetObjById(otherId)
	if !exists {
		return nil, fmt.Errorf("player with %v id doesnt exists, you are lying", otherId)
	}
	return other, nil
}
