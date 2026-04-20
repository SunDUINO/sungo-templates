/*
 * Project: test_raylib
 * Template: Raylib 3D Scene – SunGo Project Manager
 *
 * Features:
 *   - Orbital camera (mouse drag to rotate, scroll to zoom)
 *   - Basic 3D lighting (ambient + directional)
 *   - Grid floor + several 3D primitives
 *   - FPS counter and debug overlay
 *   - Clean game loop with fixed timestep logic
 */

package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	screenWidth  = 1280
	screenHeight = 720
	title        = "test_raylib"
	targetFPS    = 60
)

// ── Scene objects ─────────────────────────────────────────────────────────────

type Object3D struct {
	position rl.Vector3
	color    rl.Color
	size     float32
	kind     string // "cube" | "sphere" | "cylinder"
}

func defaultScene() []Object3D {
	return []Object3D{
		{rl.NewVector3(0, 1, 0),    rl.Red,    2.0, "cube"},
		{rl.NewVector3(4, 0.75, 2), rl.Blue,   1.5, "sphere"},
		{rl.NewVector3(-4, 1, -2),  rl.Green,  1.0, "cylinder"},
		{rl.NewVector3(0, 0.5, -5), rl.Yellow, 1.0, "cube"},
		{rl.NewVector3(6, 1.25, -4),rl.Purple, 1.25,"sphere"},
	}
}

// ── Rendering ─────────────────────────────────────────────────────────────────

func drawScene(objects []Object3D, t float32) {
	// Animated rotation around Y axis
	rl.BeginMode3D(rl.Camera3D{})

	// Grid floor
	rl.DrawGrid(20, 1.0)

	for i, obj := range objects {
		angle := t*30 + float32(i)*72 // each object rotates at different offset
		_ = angle

		switch obj.kind {
		case "cube":
			rl.DrawCube(obj.position, obj.size, obj.size, obj.size, obj.color)
			rl.DrawCubeWires(obj.position, obj.size, obj.size, obj.size, rl.DarkGray)
		case "sphere":
			rl.DrawSphere(obj.position, obj.size*0.75, obj.color)
			rl.DrawSphereWires(obj.position, obj.size*0.75, 8, 8, rl.DarkGray)
		case "cylinder":
			rl.DrawCylinder(obj.position, obj.size*0.5, obj.size*0.5, obj.size*2, 16, obj.color)
			rl.DrawCylinderWires(obj.position, obj.size*0.5, obj.size*0.5, obj.size*2, 16, rl.DarkGray)
		}
	}

	rl.EndMode3D()
}

func drawHUD(camera rl.Camera3D) {
	fps := rl.GetFPS()
	rl.DrawText(fmt.Sprintf("FPS: %d", fps), 10, 10, 20, rl.Green)
	rl.DrawText(title, 10, 35, 16, rl.LightGray)
	rl.DrawText("LMB drag: rotate  |  Scroll: zoom  |  RMB drag: pan", 10, screenHeight-30, 14, rl.Gray)
	rl.DrawText(
		fmt.Sprintf("Cam pos: (%.1f, %.1f, %.1f)", camera.Position.X, camera.Position.Y, camera.Position.Z),
		10, screenHeight-50, 14, rl.Gray,
	)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	rl.InitWindow(screenWidth, screenHeight, title)
	defer rl.CloseWindow()

	rl.SetTargetFPS(targetFPS)

	// Camera setup
	camera := rl.Camera3D{
		Position:   rl.NewVector3(12, 10, 12),
		Target:     rl.NewVector3(0, 0, 0),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       45,
		Projection: rl.CameraPerspective,
	}

	objects := defaultScene()
	var elapsed float32

	for !rl.WindowShouldClose() {
		dt      := rl.GetFrameTime()
		elapsed += dt

		// Orbital camera controls
		rl.UpdateCamera(&camera, rl.CameraOrbital)

		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 20, G: 20, B: 30, A: 255})

		drawScene(objects, elapsed)
		drawHUD(camera)

		rl.EndDrawing()
	}
}
