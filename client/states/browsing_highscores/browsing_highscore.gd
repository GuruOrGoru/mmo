extends Node

const packets := preload("res://packets.gd")

@onready var _highscores: Highscores = $MarginContainer/VBoxContainer/Highscore
@onready var _return_to_menu: Button = $MarginContainer/VBoxContainer/HBoxContainer/BackButton
@onready var _search_bar: LineEdit = $MarginContainer/VBoxContainer/HBoxContainer/SearchBar
@onready var _search_button: Button = $MarginContainer/VBoxContainer/HBoxContainer/SearchButton
@onready var _log: Log = $MarginContainer/VBoxContainer/Log

func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_received)
	_return_to_menu.pressed.connect(_on_return_to_menu_button_pressed)
	_search_bar.text_submitted.connect(_on_search_text_submitted)
	_search_button.pressed.connect(_on_search_button_pressed)

	var packet := packets.Packet.new()
	packet.new_hiscore_board_request()
	WS.send(packet)
		
func _on_return_to_menu_button_pressed() -> void:
	var packet := packets.Packet.new()
	packet.new_finished_browsing_highscore()
	WS.send(packet)
	GameManager.set_state(GameManager.State.CONNECTED)
	
func _on_search_text_submitted(_new_text: String) -> void:
	_on_search_button_pressed
	
func _on_search_button_pressed() -> void:
	var packet := packets.Packet.new()
	var hs_message := packet.new_search_highscore()
	hs_message.set_name(_search_bar.text)
	WS.send(packet)
	
func _on_ws_packet_received(packet: packets.Packet) -> void:
	if packet.has_hiscore_board():
		_handle_hiscore_board_msg(packet.get_hiscore_board())
	elif packet.has_deny_response():
		_handle_deny_response(packet.get_deny_response())

func _handle_hiscore_board_msg(hiscore_board_msg: packets.HiscoreBoardMessage) -> void:
	_highscores.clear_hiscores()
	for hiscore_msg: packets.HiscoreMessage in hiscore_board_msg.get_hiscores():
		var name := hiscore_msg.get_name()
		var rank_and_name := "%d. %s" % [hiscore_msg.get_rank(), name]
		var score := hiscore_msg.get_score()
		var highlight := name.to_lower() == _search_bar.text.to_lower()
		_highscores.set_highscore(rank_and_name, score, highlight)

func _handle_deny_response(deny_resp: packets.DenyResponseMessage) -> void:
	_log.error(deny_resp.get_reason())
	
