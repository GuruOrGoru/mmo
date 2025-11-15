class_name LoginForm
extends VBoxContainer
@onready var _username: LineEdit = $Username
@onready var _password: LineEdit = $Password
@onready var _login_button: Button = $HBoxContainer/LoginButton
@onready var _high_score: Button = $HBoxContainer/HighScore

signal form_submitted(username: String, password: String)
# Called when the node enters the scene tree for the first time.
func _ready() -> void:
	_high_score.pressed.connect(_on_highscore_button_pressed)
	_login_button.pressed.connect(_on_login_button_pressed)

func _on_highscore_button_pressed() -> void:
	GameManager.set_state(GameManager.State.BROWSINGHIGHSCORES)
	
func _on_login_button_pressed() -> void:
	form_submitted.emit(_username.text, _password.text)
