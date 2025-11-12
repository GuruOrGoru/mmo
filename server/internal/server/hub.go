package server

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"math/rand/v2"
	"net/http"

	"github.com/guruorgoru/go-mmo/server/internal/objects"
	"github.com/guruorgoru/go-mmo/server/internal/server/db"
	"github.com/guruorgoru/go-mmo/server/pkg/packets"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed db/config/schema.sql
var schemaGenSql string

type State interface {
	Name() string

	// DI the client into state handler
	SetClient(client ClientInterfacer)

	// On Entry and handling their reqs
	OnEntry()
	Handle(senderId uint64, msg packets.Msg)

	// Cleanup
	OnExit()
}

type SharedGameObjects struct {
	Players *objects.SharedCollection[*objects.Player]
	Spores  *objects.SharedCollection[*objects.Spore]
}

type ClientInterfacer interface {
	// Gives the ID of the Client
	GetId() uint64

	// Initialize a new clients by giving it id ans shi's
	Initialize(id uint64)

	// Processes the packet sent by the Client
	ProcessMessage(senderId uint64, message packets.Msg)

	// Forwards message to all other clients via Hub
	Broadcast(message packets.Msg)

	// Set and change the state of client. Eg: from login to lobby, etc
	SetState(newState State)

	// Forwards message to another client via Hub
	SendTo(message packets.Msg, senderId uint64)

	// Reads messages from Client
	ReadLoop()

	// Writes messages to Client
	WriteLoop()

	DbTx() *DbTx
	SharedGameObjects() *SharedGameObjects

	// Puts message from this client to write pump
	Send(message packets.Msg)

	// Puts message from another client to write pump
	SendAs(message packets.Msg, senderId uint64)

	// Closes the client connection
	Close(reason string)
}

// All connected clients will be managed by the Hub
type Hub struct {
	Clients *objects.SharedCollection[ClientInterfacer]

	// Packets in this channel will be broadcasted to all Clients
	BoradcastChan chan *packets.Packet

	// Clients in this channel will be registered to the Hub
	RegisterChan chan ClientInterfacer

	// Clients in this channel will be unregistered from the Hub
	UnregisterChan chan ClientInterfacer

	dbPool *sql.DB

	SharedGameObjects *SharedGameObjects
}

type DbTx struct {
	Ctx     context.Context
	Queries *db.Queries
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:     context.Background(),
		Queries: db.New(h.dbPool),
	}
}

func NewHub() *Hub {
	dbPool, err := sql.Open("pgx", "postgres://guruorgoru:balakotalu77@localhost:5432/mmo")
	if err != nil {
		log.Fatal(err)
	}
	return &Hub{
		Clients:        objects.NewSharedCollection[ClientInterfacer](0),
		BoradcastChan:  make(chan *packets.Packet),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
		dbPool:         dbPool,
		SharedGameObjects: &SharedGameObjects{
			Players: objects.NewSharedCollection[*objects.Player](0),
			Spores: objects.NewSharedCollection[*objects.Spore](0),
		},
	}
}

func (h *Hub) Run() {
	log.Println("Initializing database...")
	if _, err := h.dbPool.ExecContext(context.Background(), schemaGenSql); err != nil {
		log.Fatalln("Error initializing database", err)
	}

	log.Println("Spawning spores...")
	for range objects.SpawnLimit {
		h.SharedGameObjects.Spores.Add(h.newSpores())
	}

	log.Println("Hub is listening...")

	for {
		select {
		case client := <-h.RegisterChan:
			client.Initialize(h.Clients.Add(client))
			log.Printf("Client %v registered", client.GetId())
		case client := <-h.UnregisterChan:
			h.Clients.Delete(client.GetId())
		case packet := <-h.BoradcastChan:
			h.Clients.ForEach(func(id uint64, client ClientInterfacer) {
				if id != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error), w http.ResponseWriter, r *http.Request) {
	client, err := getNewClient(h, w, r)
	if err != nil {
		log.Println("Error while getting new client:", err)
		return
	}

	h.RegisterChan <- client

	go client.WriteLoop()
	go client.ReadLoop()
}

func (h *Hub) newSpores() *objects.Spore {
	radius := max(rand.NormFloat64()*3+15, 7)
	x, y := objects.SpawnCoords(radius, h.SharedGameObjects.Players, h.SharedGameObjects.Spores)
	return &objects.Spore{
		Radius: radius,
		X: x,
		Y: y,
	}
}
