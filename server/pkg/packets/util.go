package packets

type Msg = isPacket_Msg

func NewChat(msg string) Msg {
	return  &Packet_ChatMsg{
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
