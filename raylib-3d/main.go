/*
 * Project: ${projectName}
 * Template: Raylib 3D Scene – SunGo Project Manager
 *
 * Features:
 *   - Orbital camera (mouse drag to rotate, scroll to zoom)
 *   - Grid floor + 3D primitives (cube, sphere, cylinder)
 *   - FPS counter and debug overlay
 *
 * NOTE: BeginMode3D must always be inside BeginDrawing/EndDrawing block.
 */

package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	windowTitle  = "${projectName}"
	targetFPS    = 60
)

type Object3D struct {
	position rl.Vector3
	color    rl.Color
	size     float32
	kind     string
}

func defaultScene() []Object3D {
	return []Object3D{
		{rl.NewVector3(0, 1, 0),     rl.Red,    2.0,  "cube"},
		{rl.NewVector3(4, 0.75, 2),  rl.Blue,   1.5,  "sphere"},
		{rl.NewVector3(-4, 1, -2),   rl.Green,  1.0,  "cylinder"},
		{rl.NewVector3(0, 0.5, -5),  rl.Yellow, 1.0,  "cube"},
		{rl.NewVector3(6, 1.25, -4), rl.Purple, 1.25, "sphere"},
	}
}

func drawScene(objects []Object3D) {
	rl.DrawGrid(20, 1.0)
	for _, obj := range objects {
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
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, windowTitle)
	defer rl.CloseWindow()
	rl.SetTargetFPS(targetFPS)

	camera := rl.Camera3D{
		Position:   rl.NewVector3(12, 10, 12),
		Target:     rl.NewVector3(0, 0, 0),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       45,
		Projection: rl.CameraPerspective,
	}

	objects := defaultScene()

	for !rl.WindowShouldClose() {
		// Update camera (orbital: LMB drag = rotate, scroll = zoom)
		rl.UpdateCamera(&camera, rl.CameraOrbital)

		// ── Draw ─────────────────────────────────────────────────────────────
		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 20, G: 20, B: 30, A: 255})

		// 3D pass – must be between BeginMode3D / EndMode3D
		rl.BeginMode3D(camera)
		drawScene(objects)
		rl.EndMode3D()

		// 2D HUD – drawn after EndMode3D, still inside BeginDrawing
		rl.DrawText(fmt.Sprintf("FPS: %d", rl.GetFPS()), 10, 10, 20, rl.Green)
		rl.DrawText(windowTitle, 10, 35, 16, rl.LightGray)
		rl.DrawText(
			fmt.Sprintf("Cam: (%.1f, %.1f, %.1f)",
				camera.Position.X, camera.Position.Y, camera.Position.Z),
			10, screenHeight-50, 14, rl.Gray,
		)
		rl.DrawText(
			"LMB drag: rotate  |  Scroll: zoom  |  ESC: quit",
			10, screenHeight-28, 14, rl.Gray,
		)

		rl.EndDrawing()
	}
}
