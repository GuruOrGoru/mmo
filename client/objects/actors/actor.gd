extends Area2D

const Actor := preload("res://objects/actors/actor.gd")
const Scene := preload("res://objects/actors/actor.tscn")
const packets := preload("res://packets.gd")

var _target_zoom := 2.0
var _furthest_zoom_allowed := _target_zoom
var actor_id: int
var actor_name: String
var start_x: float
var start_y: float
var start_rad: float
var speed: float
var is_player: bool
var color: Color

var server_position: Vector2
var velocity: Vector2
var radius: float:
	set(new_radius):
		radius = new_radius
		if _collision_shape:
			_collision_shape.set_radius(radius)
		_update_zoom()
		queue_redraw()
		
@onready var _nameplate: Label = $Nameplate
@onready var _camera: Camera2D = $Camera2D
@onready var _collision_shape: CircleShape2D = $CollisionShape2D.shape

static func instantiate(actor_id: int, actor_name: String, x: float, y: float, radius: float, speed: float, is_player: bool, color: Color) -> Actor:
	var actor = Scene.instantiate()
	actor.actor_id = actor_id
	actor.actor_name = actor_name
	actor.start_x = x
	actor.start_y = y
	actor.start_rad = radius
	actor.speed = speed
	actor.is_player = is_player
	actor.color = color

	return actor

	
func _update_zoom() -> void:
	if is_node_ready():
		_nameplate.add_theme_font_size_override("font_size", max(16, radius/2))
	
	if not is_player:
		return
		
	var new_furthest_zoom_allowed := 2 * start_rad / radius
	if is_equal_approx(_target_zoom, _furthest_zoom_allowed):
		_target_zoom = new_furthest_zoom_allowed
	_furthest_zoom_allowed = new_furthest_zoom_allowed

func _ready():

	position.x = start_x
	position.y = start_y
	server_position = position
	velocity = Vector2.RIGHT * speed
	radius = start_rad
	_collision_shape.radius = radius
	_nameplate.text = actor_name

func _input(event):
	if is_player and event is InputEventMouseButton and event.is_pressed():
		match event.button_index:
			MOUSE_BUTTON_WHEEL_UP:
				_target_zoom = min(4, _target_zoom + 0.1)
			MOUSE_BUTTON_WHEEL_DOWN:
				_target_zoom = max(_furthest_zoom_allowed, _target_zoom - 0.1)


func _physics_process(delta) -> void:
	position += velocity * delta
	server_position += velocity * delta
	position += (server_position - position)*0.4

	if not is_player:
		return
	# Player-specific stuff below here
		
	var mouse_pos := get_global_mouse_position()

	var input_vec = position.direction_to(mouse_pos).normalized()
	var target_velocity = input_vec * speed
	velocity = velocity.lerp(target_velocity, 5.0 * delta)
	if velocity.length_squared() > 0.01:  # Only send if moving
		var packet := packets.Packet.new()
		var player_direction_message := packet.new_player_direction()
		player_direction_message.set_direction(velocity.angle())
		WS.send(packet)

func _draw() -> void:
	draw_circle(Vector2.ZERO, _collision_shape.radius, color)
	
func _process(_delta: float) -> void:
	if not is_equal_approx(_camera.zoom.x, _target_zoom):
		_camera.zoom -= Vector2(1, 1) * (_camera.zoom.x - _target_zoom) * 0.05
