extends Node

enum State {
	ENTERED,
	CONNECTED,
	INGAME,
}

var _state_scenes: Dictionary[State, String] = {
	State.ENTERED: "res://states/entered/entered.tscn",
	State.CONNECTED: "res://states/connected/connected.tscn",
	State.INGAME: "res://states/ingame/ingame.tscn",
}

var client_id: int
var _current_scene_root: Node

# Setting state inside the client
func set_state(state: State) -> void:
	if _current_scene_root != null:
		_current_scene_root.queue_free()
		
	var scene: PackedScene = load(_state_scenes[state])
	
	_current_scene_root = scene.instantiate()
	add_child(_current_scene_root)
