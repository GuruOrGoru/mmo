class_name RegisterForm
extends VBoxContainer

@onready var _username: LineEdit = $Username
@onready var _password: LineEdit = $Password
@onready var _comfirm_passw: LineEdit = $ComfirmPassw
@onready var _register_button: Button = $HBoxContainer/RegisterButton
@onready var _cancel: Button = $HBoxContainer/Cancel
@onready var _color_picker: ColorPicker = $ColorPicker


signal form_submitted(username: String, password: String, confirm_password: String)
signal form_cancelled()
# Called when the node enters the scene tree for the first time.
func _ready() -> void:
	_register_button.pressed.connect(_on_register_pressed)
	_cancel.pressed.connect(_on_cancel_pressed)

func _on_register_pressed() -> void:
	form_submitted.emit(_username.text, _password.text, _comfirm_passw.text, _color_picker.color)
	
func _on_cancel_pressed() -> void:
	form_cancelled.emit()
