class_name Spore
extends Area2D

const Scene := preload("res://objects/spores/spore.tscn")
const Actor := preload("res://objects/actors/actor.gd")

var spore_id: int
var x: float
var y: float
var radius: float
var color: Color
var underneathPlayer: bool

@onready var _collision_shape: CircleShape2D = $CollisionShape2D.shape

static func instantiate(spore_id: int, x: float, y: float, radius: float, underneath_player: bool) -> Spore:
	var spore := Scene.instantiate()
	spore.spore_id = spore_id
	spore.x = x
	spore.y = y
	spore.radius = radius
	spore.underneathPlayer = underneath_player
	
	return spore

func _ready() -> void:
	position.x = x
	position.y = y
	_collision_shape.radius = radius
	if underneathPlayer:
		area_exited.connect(_on_area_exited)

	color = Color.from_hsv(randf(), 1, 1, 1)

func _draw() -> void:
	draw_circle(Vector2.ZERO, radius, color)

func _on_area_exited(area: Area2D) -> void:
	if area is Actor:
		underneathPlayer = false
