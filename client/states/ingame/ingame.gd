extends Node

const packets := preload("res://packets.gd")
@onready var _log: Log = $UI/Log
@onready var _line_edit: LineEdit = $UI/LineEdit


func _ready() -> void:
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_received)
	_line_edit.text_submitted.connect(_on_line_edit_text_submitted)	
	
func _on_ws_connection_closed() -> void:
	_log.warning("Connection closed")

func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_chat_msg():
		_handle_chat_msg(sender_id, packet.get_chat_msg())


func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	_log.chat("Client %d" % sender_id, chat_msg.get_msg())


func _on_line_edit_text_submitted(text: String) -> void:
	var packet := packets.Packet.new()
	var chat_msg := packet.new_chat_msg()
	chat_msg.set_msg(text)
	
	var err := WS.send(packet)
	if err:
		_log.error("Error sending chat message")
	else:
		_log.chat("You", text)
	_line_edit.text = ""
