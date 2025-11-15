class_name Highscores
extends ScrollContainer

var _scores: Array[int]

@onready var _vbox: VBoxContainer = $MarginContainer/VBoxContainer
@onready var _entry_template: HBoxContainer = $MarginContainer/VBoxContainer/HBoxContainer
@onready var _return_to_menu: Button = $"../HBoxContainer/ReturnToMenu"
@onready var _highscores: Highscores = $"."

func _ready() -> void:
	_entry_template.hide()

func set_highscore(name: String, score: int, highlight: bool = false) -> void:
	remove_highscore(name)
	_add_highscore(name, score, highlight)
	
func _add_highscore(name: String, score: int, highlight: bool) -> void:
	_scores.append(score)
	_scores.sort()
	var pos := len(_scores) - _scores.find(score) - 1 # -1 to keep real entries above the template
	
	var entry: HBoxContainer = _entry_template.duplicate()
	var name_label: Label = entry.get_child(0)
	var score_label: Label = entry.get_child(1)
	
	_vbox.add_child(entry)
	
	_vbox.move_child(entry, pos)
	
	name_label.text = name
	score_label.text = str(score)
	
	if highlight:
		name_label.add_theme_color_override("font_color", Color.LIGHT_SALMON)
	
	entry.show()


func remove_highscore(name: String) -> void:
	for i in range(len(_scores)):
		var entry: HBoxContainer = _vbox.get_child(i)
		var name_label: Label = entry.get_child(0)

		if name_label.text == name:
			_scores.remove_at(len(_scores) - i - 1)

			entry.free()
			return
			
func clear_hiscores() -> void:
	_scores.clear()
	for entry in _vbox.get_children():
		if entry != _entry_template:
			entry.free()
