package clients

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/guruorgoru/go-mmo/server/internal/server"
	"github.com/guruorgoru/go-mmo/server/pkg/packets"
	"google.golang.org/protobuf/proto"
)

type WebSocketClient struct {
	id       uint64
	conn     *websocket.Conn
	hub      *server.Hub
	sendChan chan *packets.Packet
	logger   *log.Logger
}

func NewWebSocketClient(h *server.Hub, w http.ResponseWriter, r *http.Request) (server.ClientInterfacer, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(_ *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		hub:      h,
		conn:     conn,
		sendChan: make(chan *packets.Packet, 256),
		logger:   log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
	}

	return client, nil
}

// Interface methods for client <start>

func (c *WebSocketClient) GetId() uint64 {
	return c.id
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprintf("Client %v", c.id))
	c.SendMessage(packets.NewId(c.id))
	c.logger.Println("Sent Id to client")
}

func (c *WebSocketClient) ProcessMessage(senderId uint64, msg packets.Msg) {
	if senderId == c.id {
		c.BroadcastMessage(msg)
	} else {
		c.SendMessageAs(msg, senderId)
	}
}

func (c *WebSocketClient) SendMessage(message packets.Msg) {
	c.SendMessageAs(message, c.id)
}

func (c *WebSocketClient) SendMessageAs(message packets.Msg, senderId uint64) {
	select {
	case c.sendChan <- &packets.Packet{SenderId: senderId, Msg: message}:
	default:
		c.logger.Printf("Client %d send channel full, dropping message: %T", c.id, message)
	}
}

func (c *WebSocketClient) PassToPeer(msg packets.Msg, peerId uint64) {
	if peer, exists := c.hub.Clients.GetObjById(peerId); exists {
		peer.ProcessMessage(c.id, msg)
	}
}

func (c *WebSocketClient) BroadcastMessage(msg packets.Msg) {
	c.hub.BoradcastChan <- &packets.Packet{SenderId: c.id, Msg: msg}
}

func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.logger.Println("Closing read pump")
		c.Close("read pump closed")
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("error: %v", err)
			}
			break
		}

		packet := &packets.Packet{}
		err = proto.Unmarshal(data, packet)
		if err != nil {
			c.logger.Printf("error unmarshalling data: %v", err)
			continue
		}

		// To allow the client to lazily not set the sender ID, we'll assume they want to send it as themselves
		if packet.SenderId == 0 {
			packet.SenderId = c.id
		}

		c.ProcessMessage(packet.SenderId, packet.Msg)
	}
}

func (c *WebSocketClient) WritePump() {
	defer func() {
		c.logger.Println("Closing write pump")
		c.Close("write pump closed")
	}()

	for packet := range c.sendChan {
		writer, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			c.logger.Printf("error getting writer for %T packet, closing client: %v", packet.Msg, err)
			return
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			c.logger.Printf("error marshalling %T packet, dropping: %v", packet.Msg, err)
			continue
		}

		_, writeErr := writer.Write(data)

		if writeErr != nil {
			c.logger.Printf("error writing %T packet: %v", packet.Msg, err)
			continue
		}

		_, err = writer.Write([]byte{'\n'})
		if err != nil {
			c.logger.Fatalf("error writing the data %v", err)
			continue
		}

		if closeErr := writer.Close(); closeErr != nil {
			c.logger.Printf("error closing writer, dropping %T packet: %v", packet.Msg, err)
			continue
		}
	}
}

func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client connection because: %s", reason)

	c.hub.UnregisterChan <- c
	err := c.conn.Close()
	if err != nil {
		c.logger.Fatalln(err)
		return
	}
	if _, closed := <-c.sendChan; !closed {
		close(c.sendChan)
	}
}
