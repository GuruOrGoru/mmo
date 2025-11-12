package packets

import "github.com/guruorgoru/go-mmo/server/internal/objects"

type Msg = isPacket_Msg

func NewChat(msg string) Msg {
	return &Packet_ChatMsg{
		ChatMsg: &ChatMessage{
			Msg: msg,
		},
	}
}

func NewId(id uint64) Msg {
	return &Packet_IdMsg{
		IdMsg: &IdMessage{
			Id: id,
		},
	}
}

func NewOkMessage() Msg {
	return &Packet_OkResponse{
		OkResponse: &OkResponseMessage{},
	}
}

func NewDenyMessage(reason string) Msg {
	return &Packet_DenyResponse{
		DenyResponse: &DenyResponseMessage{
			Reason: reason,
		},
	}
}

func NewPlayer(id uint64, player *objects.Player) Msg {
	return &Packet_Player{
		Player: &PlayerMessage{
			Id:        id,
			Name:      player.Name,
			X:         player.X,
			Y:         player.Y,
			Radius:    player.Radius,
			Direction: player.Direction,
			Speed:     player.Speed,
		},
	}
}

func NewSpore(id uint64, spore *objects.Spore) Msg {
	return &Packet_Spore{
		Spore: newSporeMessage(id, spore),
	}
}

func NewSporeBatch(spores map[uint64]*objects.Spore) Msg {
	sporeMessages := make([]*SporeMessage, 0, len(spores))

	for id, obj := range spores {
		sporeMessages = append(sporeMessages, newSporeMessage(id, obj))
	}

	return &Packet_SporeBatch{
		SporeBatch: &SporesBatchMessage{
			Spores: sporeMessages,
		},
	}
}

func newSporeMessage(id uint64, spore *objects.Spore) *SporeMessage {
	return &SporeMessage{
		Id:     id,
		X:      spore.X,
		Y:      spore.Y,
		Radius: spore.Radius,
	}
}
