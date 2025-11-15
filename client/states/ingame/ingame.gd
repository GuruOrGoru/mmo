extends Node

const packets := preload("res://packets.gd")
const Actor := preload("res://objects/actors/actor.gd")
const Spore := preload("res://objects/spores/spore.gd")

@onready var _log: Log = $UI/MarginContainer/VBoxContainer/Log
@onready var _line_edit: LineEdit = $UI/MarginContainer/VBoxContainer/HBoxContainer/LineEdit
@onready var _highscore: Highscores = $UI/HighScoresBox/Highscore
@onready var _send_button: Button = $UI/MarginContainer/VBoxContainer/HBoxContainer/SendButton
@onready var _logout_button: Button = $UI/MarginContainer/VBoxContainer/HBoxContainer/LogoutButton
@onready var _world: Node2D = $World
@onready var _coords: Label = $UI/coords
var _players: Dictionary[int, Actor]
var _spores: Dictionary[int, Spore]
var _target_zoom := 2
var _furthest_zoom_allowed := _target_zoom

func _ready() -> void:
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_received)
	_send_button.pressed.connect(_on_send_button_pressed)
	_logout_button.pressed.connect(_on_logout_button_pressed)
	_line_edit.text_submitted.connect(_on_line_edit_text_submitted)	

func _on_logout_button_pressed() -> void:
	var packet := packets.Packet.new()
	var disMsg := packet.new_disconnect()
	disMsg.set_reason("they logged out")
	WS.send(packet)
	
	GameManager.set_state(GameManager.State.CONNECTED)

func _on_send_button_pressed() -> void:
	_on_line_edit_text_submitted(_line_edit.text)

func _process(delta: float) -> void:
	if GameManager.client_id in _players:
		var player = _players[GameManager.client_id]
		_coords.text = "x: %.1f, y: %.1f" % [player.position.x, player.position.y]
func _on_ws_connection_closed() -> void:
	_log.warning("Connection closed")

func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_chat_msg():
		_handle_chat_msg(sender_id, packet.get_chat_msg())
	elif packet.has_player():
		_handle_player_msg(sender_id, packet.get_player())
	elif packet.has_spore():
		_handle_spore_msg(sender_id, packet.get_spore())
	elif packet.has_spore_batch():
		_handle_spore_batch(sender_id, packet.get_spore_batch())
	elif packet.has_spore_consumed():
		_handle_spore_consumed(sender_id, packet.get_spore_consumed())
	elif  packet.has_disconnect():
		_handle_disconnect_msg(sender_id, packet.get_disconnect())

func _handle_disconnect_msg(sender_id: int, msg: packets.DisconnectMessage) -> void:
	if sender_id in _players:
		var player := _players[sender_id]
		var reason := msg.get_reason()
		_log.info("%s disconnected because %s" % [player.actor_name, reason])
		_remove_actor(player)
	
func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	if sender_id in _players:
		var actor := _players[sender_id]
		_log.chat(actor.actor_name, chat_msg.get_msg())
		
func _handle_player_msg(sender_id: int, player_msg: packets.PlayerMessage) -> void:
	var actor_id := player_msg.get_id()
	var actor_name := player_msg.get_name()
	var x := player_msg.get_x()
	var y := player_msg.get_y()
	var radius := player_msg.get_radius()
	var speed := player_msg.get_speed()
	var color_msg := player_msg.get_color()
	var color := Color.hex(color_msg)

	var is_player := actor_id == GameManager.client_id

	if actor_id not in _players:
		_add_actor(actor_id, actor_name, x, y, radius, speed, is_player, color)
	else:
		var direction := player_msg.get_direction()
		_update_actor(actor_id, x, y, direction, speed, radius, is_player)

func _add_actor(actor_id: int, actor_name: String, x: float, y: float, radius: float, speed: float, is_player: bool, color: Color) -> void:
	var actor := Actor.instantiate(actor_id, actor_name, x, y, radius, speed, is_player, color)
	_set_actor_mass(actor, _rad_to_mass(radius))
	_world.add_child(actor)
	actor.z_index = 1
	_players[actor_id] = actor
	
	if is_player:
		actor.area_entered.connect(_on_player_area_entered)

func _update_actor(actor_id: int, x: float, y: float, direction: float, speed: float, radius: float, is_player: bool) -> void:
	var actor := _players[actor_id]
	_set_actor_mass(actor, _rad_to_mass(radius))

	if is_player or actor.position.distance_squared_to(Vector2(x, y)) > 100:
		actor.server_position.x = x
		actor.server_position.y = y
	
	if not is_player:
		actor.velocity = Vector2.from_angle(direction) * speed

func _on_player_area_entered(area: Area2D) -> void:
	if area is Spore:
		_consume_spore(area as Spore)
	if area is Actor:
		_collide_actor(area as Actor)
		
func _collide_actor(actor: Actor) -> void:
	var player := _players[GameManager.client_id]
	var player_mass := _rad_to_mass(player.radius)
	var actor_mass := _rad_to_mass(actor.radius)
	
	if player_mass > actor_mass * 1.2:
		_consume_player(actor)

func _consume_player(actor: Actor) -> void:
	var player := _players[GameManager.client_id]
	var player_mass := _rad_to_mass(player.radius)
	var actor_mass := _rad_to_mass(actor.radius)
	_set_actor_mass(player, player_mass + actor_mass)
	
	var packet := packets.Packet.new()
	var player_consumed_msg := packet.new_player_consumed()
	player_consumed_msg.set_player_id(actor.actor_id)
	WS.send(packet)
	_remove_actor(actor)

func _consume_spore(spore: Spore) -> void:
	var player = _players[GameManager.client_id]
	if spore.underneathPlayer:
		return
	var player_mass := _rad_to_mass(player.radius)
	var spore_mass := _rad_to_mass(spore.radius)
	_set_actor_mass(player, player_mass + spore_mass)
	var packet := packets.Packet.new()
	var spore_consumed_msg := packet.new_spore_consumed()
	spore_consumed_msg.set_spore_id(spore.spore_id)
	WS.send(packet)
	_remove_spore(spore)

func _remove_spore(spore: Spore) -> void:
	_spores.erase(spore.spore_id)
	spore.queue_free()

func _remove_actor(actor: Actor) -> void:
	_players.erase(actor.actor_id)
	_highscore.remove_highscore(actor.actor_name)
	actor.queue_free()

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
	
func _handle_spore_msg(sender_id: int, spore_msg: packets.SporeMessage) -> void:
	var spore_id := spore_msg.get_id()
	var x := spore_msg.get_x()
	var y := spore_msg.get_y()
	var radius := spore_msg.get_radius()
	var underneathPlayer := false

	if GameManager.client_id in _players:
		var player := _players[GameManager.client_id]
		var player_pos := Vector2(player.position.x, player.position.y)
		var spore_pos := Vector2(x, y)
		underneathPlayer = player_pos.distance_squared_to(spore_pos) < player.radius * player.radius

	if spore_id not in _spores:
		var spore := Spore.instantiate(spore_id, x, y, radius, underneathPlayer)
		_world.add_child(spore)
		_spores[spore_id] = spore
func _handle_spore_batch(sender_id: int, spore_batch_msg: packets.SporesBatchMessage) -> void:
	for spore_msg in spore_batch_msg.get_spores():
		_handle_spore_msg(sender_id, spore_msg)

func  _handle_spore_consumed(sender_id: int, spore_consumed_msg: packets.SporeConsumedMessage) -> void:
	if sender_id in _players:
		var actor := _players[sender_id]
		var actor_mass := _rad_to_mass(actor.radius)

		var spore_id := spore_consumed_msg.get_spore_id()
		if spore_id in _spores:
			var spore := _spores[spore_id]
			var spore_mass := _rad_to_mass(spore.radius)
			_set_actor_mass(actor, actor_mass + spore_mass)
			_remove_spore(spore)

func _rad_to_mass(radius: float) -> float:
	return radius * radius * PI
func _set_actor_mass(actor: Actor, new_mass: float) -> void:
	actor.radius = sqrt(new_mass / PI)
	_highscore.set_highscore(actor.actor_name, new_mass)
