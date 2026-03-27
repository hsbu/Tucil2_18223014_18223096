package main

import (
	"image/color"
	"math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	viewWidth         = 900
	viewHeight        = 700
	mouseSensitivity  = 0.005
	scrollSensitivity = 0.9
	focalLengthRatio  = 0.9
	minCameraDistance = 0.01
)

// Camera defines an orbital camera system utilizing spherical coordinates
type Camera struct {
	Yaw      float64 // Horizontal rotation 
	Pitch    float64 // Vertical rotation 
	Distance float64 // Radius from the target point
	Target   Vertex  // The focal point the camera orbits around
}

// VoxelViewer maintains the global state of the 3D rendering scene
type VoxelViewer struct {
	voxels     []AABB
	camera     Camera
	lastMouseX int
	lastMouseY int
	isDragging bool
}

//Vector Operations
// norm computes the unit vector
func norm(v Vertex) Vertex {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if magnitude < 1e-10 { 
		return v
	}
	return Vertex{v.X / magnitude, v.Y / magnitude, v.Z / magnitude}
}

// cross computes the cross product
func cross(a, b Vertex) Vertex {
	return Vertex{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

// dot computes the dot product
func dot(a, b Vertex) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// subtract calculates the directional vector from b to a
func subtract(a, b Vertex) Vertex {
	return Vertex{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

//Camera 
// position the camera in x, y, z direction
func (cam *Camera) position() Vertex {
	return Vertex{
		X: cam.Target.X + cam.Distance*math.Cos(cam.Pitch)*math.Sin(cam.Yaw),
		Y: cam.Target.Y + cam.Distance*math.Sin(cam.Pitch),
		Z: cam.Target.Z + cam.Distance*math.Cos(cam.Pitch)*math.Cos(cam.Yaw),
	}
}

// camera movement
func (cam *Camera) basis() (right, up, forward Vertex) {
	camPos := cam.position()
	
	forward = norm(subtract(cam.Target, camPos))
	worldUp := Vertex{0, 1, 0}
	
	right = norm(cross(forward, worldUp))
	up = cross(right, forward) 
	
	return right, up, forward
}

//3D to 2D Projection
func (viewer *VoxelViewer) project(point Vertex) (screenX float32, screenY float32, isVisible bool) {
	right, up, forward := viewer.camera.basis()
	
	// Compute the direction and distance vector relative to the camera
	localPos := subtract(point, viewer.camera.position())

	camX := dot(localPos, right)
	camY := dot(localPos, up)
	depth := dot(localPos, forward) 

	if depth <= 0.001 {
		return 0, 0, false
	}

	focalLength := float64(viewHeight) * focalLengthRatio
	
	screenX = float32(camX/depth*focalLength) + float32(viewWidth)/2
	screenY = float32(-camY/depth*focalLength) + float32(viewHeight)/2 
	
	return screenX, screenY, true
}

//Geometry Construction 
func boxCorners(box AABB) [8]Vertex {
	return [8]Vertex{
		{box.Min.X, box.Min.Y, box.Min.Z}, 
		{box.Max.X, box.Min.Y, box.Min.Z}, 
		{box.Min.X, box.Max.Y, box.Min.Z}, 
		{box.Max.X, box.Max.Y, box.Min.Z}, 
		{box.Min.X, box.Min.Y, box.Max.Z}, 
		{box.Max.X, box.Min.Y, box.Max.Z}, 
		{box.Min.X, box.Max.Y, box.Max.Z}, 
		{box.Max.X, box.Max.Y, box.Max.Z}, 
	}
}

var boxEdges = [12][2]int{
	{0, 1}, {2, 3}, {4, 5}, {6, 7}, 
	{0, 2}, {1, 3}, {4, 6}, {5, 7}, 
	{0, 4}, {1, 5}, {2, 6}, {3, 7}, 
}

//Ebiten Engine Interface
func (viewer *VoxelViewer) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if viewer.isDragging {
			deltaX := mouseX - viewer.lastMouseX
			deltaY := mouseY - viewer.lastMouseY
			
			viewer.camera.Yaw += float64(deltaX) * mouseSensitivity
			viewer.camera.Pitch -= float64(deltaY) * mouseSensitivity

			// Clamp pitch to prevent the camera from flipping over the poles
			const maxPitchLimit = math.Pi/2 - 0.05
			if viewer.camera.Pitch > maxPitchLimit {
				viewer.camera.Pitch = maxPitchLimit
			} else if viewer.camera.Pitch < -maxPitchLimit {
				viewer.camera.Pitch = -maxPitchLimit
			}
		}
		viewer.isDragging = true
		viewer.lastMouseX = mouseX
		viewer.lastMouseY = mouseY
	} else {
		viewer.isDragging = false
	}

	_, scrollY := ebiten.Wheel()
	viewer.camera.Distance *= math.Pow(scrollSensitivity, scrollY)
	
	if viewer.camera.Distance < minCameraDistance {
		viewer.camera.Distance = minCameraDistance
	}

	return nil
}

// Draw renders the voxel wireframes onto the screen
func (viewer *VoxelViewer) Draw(screen *ebiten.Image) {
	backgroundColor := color.RGBA{15, 15, 25, 255}
	wireframeColor := color.RGBA{80, 200, 120, 200}
	
	screen.Fill(backgroundColor)

	for _, box := range viewer.voxels {
		corners := boxCorners(box)

		var projectedPoints [8][2]float32
		var isVisible [8]bool
		
		// Project all 8 vertices of the bounding box
		for i, corner := range corners {
			screenX, screenY, visible := viewer.project(corner)
			projectedPoints[i] = [2]float32{screenX, screenY}
			isVisible[i] = visible
		}

		// Render the connected edges
		for _, edge := range boxEdges {
			startIdx, endIdx := edge[0], edge[1]
			
			// Only stroke the line if both vertices are in front of the camera
			if isVisible[startIdx] && isVisible[endIdx] {
				vector.StrokeLine(
					screen,
					projectedPoints[startIdx][0], projectedPoints[startIdx][1],
					projectedPoints[endIdx][0], projectedPoints[endIdx][1],
					1, wireframeColor, false,
				)
			}
		}
	}

	ebitenutil.DebugPrint(screen, "Left drag: rotate  |  Scroll: zoom")
}

func (viewer *VoxelViewer) Layout(outsideWidth, outsideHeight int) (int, int) {
	return viewWidth, viewHeight
}

//Entry Point
func runViewer(voxels []AABB, bounds AABB) {
	
	// Center the camera focus on the midpoint of the bounding volume
	center := Vertex{
		(bounds.Min.X + bounds.Max.X) * 0.5,
		(bounds.Min.Y + bounds.Max.Y) * 0.5,
		(bounds.Min.Z + bounds.Max.Z) * 0.5,
	}
	modelWidth := bounds.Max.X - bounds.Min.X

	ebiten.SetWindowSize(viewWidth, viewHeight)
	ebiten.SetWindowTitle("Voxel Viewer — 3D Wireframe Rendering")

	viewer := &VoxelViewer{
		voxels: voxels,
		camera: Camera{
			Yaw:      0.6,
			Pitch:    0.4,
			Distance: modelWidth * 2, 
			Target:   center,
		},
	}

	if err := ebiten.RunGame(viewer); err != nil {
		panic(err)
	}
}