extends Node

const packets := preload("res://packets.gd")

var _action_on_ok_received: Callable

@onready var _login_form_tscn: LoginForm = $UI/MarginContainer/VBoxContainer/LoginForm_tscn
@onready var _register_form: RegisterForm = $UI/MarginContainer/VBoxContainer/RegisterForm
@onready var _log: Log = $UI/MarginContainer/VBoxContainer/Log
@onready var _registering_prompt: RichTextLabel = $UI/MarginContainer/VBoxContainer/RegisteringPrompt

func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_received)
	WS.connection_closed.connect(_on_ws_connection_closed)
	_register_form.hide()
	_login_form_tscn.form_submitted.connect(_on_login_form_submitted)
	_register_form.form_cancelled.connect(_on_register_form_cancelled)
	_register_form.form_submitted.connect(_on_register_from_submitted)
	_registering_prompt.meta_clicked.connect(_handle_register_prompt)

func _handle_register_prompt(meta) -> void:
	if meta is String and meta == "register":
		_login_form_tscn.hide()
		_registering_prompt.hide()
		_register_form.show()

func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_deny_response():
		var deny_response_message := packet.get_deny_response()
		_log.error(deny_response_message.get_reason())
	elif packet.has_ok_response():
		_action_on_ok_received.call()
	
func _on_register_form_cancelled() -> void:
	_register_form.hide()
	_login_form_tscn.show()
	_registering_prompt.show()
	
func _on_register_from_submitted(usrname: String, passw: String, conf_passw: String, color: Color) -> void:
	if conf_passw != passw:
		_log.error("Passwords don't match")
		return
		
	var pkt := packets.Packet.new()
	var register_req := pkt.new_register_request()
	register_req.set_username(usrname)
	register_req.set_password(passw)
	register_req.set_color(color.to_rgba32())
	WS.send(pkt)
	_action_on_ok_received = func(): _log.success("Account registered successfully! Now log into the account and enjoy the game!")
func _on_ws_connection_closed() -> void:
	pass
	
func _on_login_form_submitted(username: String, passw: String) -> void:
	var pkt := packets.Packet.new()
	var login_req := pkt.new_login_request()
	login_req.set_username(username)
	login_req.set_password(passw)
	WS.send(pkt)
	_action_on_ok_received = func(): GameManager.set_state(GameManager.State.INGAME)
