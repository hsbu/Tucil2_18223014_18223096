package main

import(
	"os"
	"bufio"
	"fmt"
	"strings"
	"strconv"
	"math"
	"path/filepath"
	"time"
)

type Vertex struct{
	X, Y, Z float64
}

// Representing faces of mesh
type Triangle struct{
	A, B, C Vertex // Vertices
}

type Mesh struct {
	ListOfVertex []Vertex
	Faces []Triangle
}

// Representing the space that wraps each node 
type AABB struct{
	Min, Max Vertex // AABB defined by computing 2 opposite vertex of mesh
}

// Representing the Tree Node
type TreeNode struct{
	Bounds AABB
	Children [8] * TreeNode
	IsVoxel bool
}

type Stats struct {
    NodesAtDepth []int
    PrunedAtDepth []int
    MaxDepth int
}

// Parsing the object
func parse(path string) (*Mesh, error){
	f, err := os.Open(path)
	
	if err != nil {
		return nil, err
	}
	
	defer f.Close()

	var listOfVertex []Vertex
	var faces []Triangle

	scanner := bufio.NewScanner(f)
	lineCount := 0

	for scanner.Scan(){
		lineCount += 1

		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)

		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		switch fields[0]{
			case "v":
				if len(fields) != 4{ // Input validation to check if each line contain 4 item
					return nil, fmt.Errorf("Input tidak valid pada vertex ke %v!", lineCount)
				}

				v1, e1 := strconv.ParseFloat(fields[1], 64) 
				v2, e2 := strconv.ParseFloat(fields[2], 64) 
				v3, e3 := strconv.ParseFloat(fields[3], 64) 
				
				if e1 != nil || e2 != nil || e3 != nil{
					return nil, fmt.Errorf("Titik koordinat tidak valid pada vertex ke %v!", lineCount)
				}

				listOfVertex = append(listOfVertex, Vertex{v1, v2, v3})

			case "f":
				if len(fields) != 4{
					return nil, fmt.Errorf("Input tidak valid pada face ke %v!", lineCount)
				}

				vertex1, e1 := strconv.Atoi(fields[1]) 
				vertex2, e2 := strconv.Atoi(fields[2]) 
				vertex3, e3 := strconv.Atoi(fields[3]) 

				if e1 != nil || e2 != nil || e3 != nil{
					return nil, fmt.Errorf("Vertex tidak valid pada face ke %v!", lineCount)
				}

				faces = append(faces, Triangle{A: listOfVertex[vertex1 - 1], B: listOfVertex[vertex2 - 1], C: listOfVertex[vertex3 - 1]})

			default:
				continue
		}
	}
	// Scanning error validation
	err = scanner.Err()
	if err != nil{
		return nil, err
	}

	// Vertices and faces validation
	if len(listOfVertex) == 0 {
		return nil, fmt.Errorf("Tidak ditemukan vertices pada %s", path)
	}
	if len(faces) == 0 {
		return nil, fmt.Errorf("Tidak ditemukan faces %s", path)
	}

	return &Mesh{ListOfVertex: listOfVertex, Faces: faces}, nil
}

// Forming the AABB box  
func makeBox(mesh *Mesh) AABB{
	inf := math.Inf(1)
	box := AABB{Min: Vertex{inf, inf, inf}, Max: Vertex{-inf, -inf, -inf}}
	
	// Wrap the mesh by defining the max and min vertex in 3 dimention
	for _, v := range mesh.ListOfVertex {
		if v.X < box.Min.X {
			box.Min.X = v.X
		}
		if v.Y < box.Min.Y {
			box.Min.Y = v.Y
		}
		if v.Z < box.Min.Z {
			box.Min.Z = v.Z
		}
		if v.X > box.Max.X {
			box.Max.X = v.X
		}
		if v.Y > box.Max.Y {
			box.Max.Y = v.Y
		}
		if v.Z > box.Max.Z {
			box.Max.Z = v.Z
		}
	}

	// Uniforming the wrap

	xLength := box.Max.X - box.Min.X
    yLength := box.Max.Y - box.Min.Y
    zLength := box.Max.Z - box.Min.Z

	// Choosing the highest length so all parts of mesh are covered
    side := math.Max(xLength, math.Max(yLength, zLength))
    
	xCenter := (box.Min.X + box.Max.X) * 0.5
    yCenter := (box.Min.Y + box.Max.Y) * 0.5
    zCenter := (box.Min.Z + box.Max.Z) * 0.5
    
	// Forming the box from the center point
	return AABB{Min: Vertex{xCenter - (0.5 * side), yCenter - (0.5 * side), zCenter - (0.5 * side)}, Max: Vertex{xCenter + (0.5 * side), yCenter + (0.5 * side), zCenter + (0.5 * side)}}
}

// Checking if the box wraps intersect with mesh using SAT Theorem
func checkIntersections(triangle Triangle, box AABB) bool{
	// Defining the new coordinates for triangle 
	xCenter := (box.Min.X + box.Max.X) * 0.5
	yCenter := (box.Min.Y + box.Max.Y) * 0.5
	zCenter := (box.Min.Z + box.Max.Z) * 0.5

	xHalf := (box.Max.X - box.Min.X) * 0.5
	yHalf := (box.Max.Y - box.Min.Y) * 0.5
	zHalf := (box.Max.Z - box.Min.Z) * 0.5

	// Shift triangle so center of triangle is at the center of box
	newVertex1 := Vertex{triangle.A.X - xCenter, triangle.A.Y - yCenter, triangle.A.Z - zCenter}
	newVertex2 := Vertex{triangle.B.X - xCenter, triangle.B.Y - yCenter, triangle.B.Z - zCenter}
	newVertex3 := Vertex{triangle.C.X - xCenter, triangle.C.Y - yCenter, triangle.C.Z - zCenter}

	// Compute the direction vectors of the triangle
	edge1 := Vertex{newVertex2.X - newVertex1.X, newVertex2.Y - newVertex1.Y, newVertex2.Z - newVertex1.Z} // New Vertex 1 -> 2
	edge2 := Vertex{newVertex3.X - newVertex2.X, newVertex3.Y - newVertex2.Y, newVertex3.Z - newVertex2.Z} 
	edge3 := Vertex{newVertex1.X - newVertex3.X, newVertex1.Y - newVertex3.Y, newVertex1.Z - newVertex3.Z}

	// isSeparated checks if an axis has a gap between the triangle and the box
	// Projects all triangle vertices onto the axis, computes the box projection radius, and checks if the triangle shadow falls outside of the box radius
	isSeparated := func(ax, ay, az float64) bool {
		p0 := ax * newVertex1.X + ay * newVertex1.Y + az * newVertex1.Z
		p1 := ax * newVertex2.X + ay * newVertex2.Y + az * newVertex2.Z
		p2 := ax * newVertex3.X + ay * newVertex3.Y + az * newVertex3.Z
		r := xHalf * math.Abs(ax) + yHalf * math.Abs(ay) + zHalf * math.Abs(az)
		
		if math.Max(p0, math.Max(p1, p2)) < -r || math.Min(p0, math.Min(p1, p2)) > r{
			return true
		} else {
			return false
		}
	}

	// Cross product of each triangle edge with each box edge direction (x, y, z)
	for _, e := range [3]Vertex{edge1, edge2, edge3} {
		if isSeparated(0, e.Z, -e.Y) {
			return false
		}
		if isSeparated(-e.Z, 0, e.X){
			return false
		}
		if isSeparated(e.Y, -e.X, 0){
			return false
		}
	}

	// Test 3 box face normals
	if isSeparated(1, 0, 0){
		return false
	}
	if isSeparated(0, 1, 0){
		return false
	}
	if isSeparated(0, 0, 1){
		return false
	}

	// Test the triangle face normal
	triangleNormal := Vertex{edge1.Y * edge2.Z - edge1.Z * edge2.Y, edge1.Z * edge2.X - edge1.X * edge2.Z, edge1.X * edge2.Y - edge1.Y * edge2.X}
	if isSeparated(triangleNormal.X, triangleNormal.Y, triangleNormal.Z){ 
		return false
	}	
	return true
}

func makeTree(bounds AABB, triangles []Triangle, depth, maxDepth int, stats *Stats) *TreeNode{
	// Store triangle that intersects with AABB box
	intersects := triangles[:0:0]
	for _, tri := range triangles {
		if checkIntersections(tri, bounds) {
			intersects = append(intersects, tri)
		}
	}

	// If the node is empty, delete
	if len(intersects) == 0 {
		stats.PrunedAtDepth[depth] += 1
		return nil
	}
	stats.NodesAtDepth[depth] += 1

	// If this is the max depth, return node as a voxel
	if depth == maxDepth {
		return &TreeNode{Bounds: bounds, IsVoxel: true}
	}

	xCenter := (bounds.Min.X + bounds.Max.X) * 0.5
	yCenter := (bounds.Min.Y + bounds.Max.Y) * 0.5
	zCenter := (bounds.Min.Z + bounds.Max.Z) * 0.5

	// Form the children from the min, center, and max point of the box
	octants := [8]AABB{
		{bounds.Min, Vertex{xCenter, yCenter, zCenter}},
		{Vertex{xCenter, bounds.Min.Y, bounds.Min.Z}, Vertex{bounds.Max.X, yCenter, zCenter}},
		{Vertex{bounds.Min.X, yCenter, bounds.Min.Z}, Vertex{xCenter, bounds.Max.Y, zCenter}},
		{Vertex{xCenter, yCenter, bounds.Min.Z}, Vertex{bounds.Max.X, bounds.Max.Y, zCenter}},
		{Vertex{bounds.Min.X, bounds.Min.Y, zCenter}, Vertex{xCenter, yCenter, bounds.Max.Z}},
		{Vertex{xCenter, bounds.Min.Y, zCenter}, Vertex{bounds.Max.X, yCenter, bounds.Max.Z}},
		{Vertex{bounds.Min.X, yCenter, zCenter}, Vertex{xCenter, bounds.Max.Y, bounds.Max.Z}},
		{Vertex{xCenter, yCenter, zCenter}, bounds.Max}}

	node := &TreeNode{Bounds: bounds}
	for i, oct := range octants {
		node.Children[i] = makeTree(oct, intersects, depth + 1, maxDepth, stats)
	}
	return node
}

func getVoxels(node *TreeNode) []AABB{
	if node == nil{
		return nil
	}

	if node.IsVoxel{
		return []AABB{node.Bounds}
	}

	var allVoxels []AABB
	for _, child := range node.Children{
		allVoxels = append(allVoxels, getVoxels(child)...)
	}

	return allVoxels
}

func writeOutput(voxels []AABB, path string) (int, int, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	totalVertices := 0
	totalFaces := 0

	// Write every poxels
	for _, box := range voxels{
		// Write the vertices
		fmt.Fprintf(w, "v %g %g %g\n", box.Min.X, box.Min.Y, box.Min.Z) // 0 LBF
		fmt.Fprintf(w, "v %g %g %g\n", box.Max.X, box.Min.Y, box.Min.Z) // 1 RBF
		fmt.Fprintf(w, "v %g %g %g\n", box.Min.X, box.Max.Y, box.Min.Z) // 2 LTF
		fmt.Fprintf(w, "v %g %g %g\n", box.Max.X, box.Max.Y, box.Min.Z) // 3 RTF
		fmt.Fprintf(w, "v %g %g %g\n", box.Min.X, box.Min.Y, box.Max.Z) // 4 LBB
		fmt.Fprintf(w, "v %g %g %g\n", box.Max.X, box.Min.Y, box.Max.Z) // 5 RBB
		fmt.Fprintf(w, "v %g %g %g\n", box.Min.X, box.Max.Y, box.Max.Z) // 6 LTB
		fmt.Fprintf(w, "v %g %g %g\n", box.Max.X, box.Max.Y, box.Max.Z) // 7 RTB

		b := totalVertices + 1 // OBJ index starts with 1

		// Write the faces
		faces := [12][3]int{
			{b + 0, b + 4, b + 5}, 
			{b + 0, b + 5, b + 1},
			{b + 2, b + 3, b + 7}, 
			{b + 2, b + 7, b + 6},
			{b + 0, b + 1, b + 3},
			{b + 0, b + 3, b + 2},
			{b + 4, b + 6, b + 7},
			{b + 4, b + 7, b + 5},
			{b + 0, b + 2, b + 6}, 
			{b + 0, b + 6, b + 4}, 
			{b + 1, b + 5, b + 7}, 
			{b + 1, b + 7, b + 3}}

		for _, face := range faces {
			fmt.Fprintf(w, "f %d %d %d\n", face[0], face[1], face[2])
		}

		totalVertices += 8
		totalFaces += 12
	}

	return totalVertices, totalFaces, w.Flush()
}

func main(){
	// Input validation
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Fprintln(os.Stderr, "Format input: tucil2 <input.obj> <max_depth>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	maxDepth, err := strconv.Atoi(os.Args[2])
	if err != nil || maxDepth < 1 {
		fmt.Fprintln(os.Stderr, "Error: tingkat kedalaman harus bernilai positif!")
		os.Exit(1)
	}

	start := time.Now()

	mesh, err := parse(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	bounds := makeBox(mesh)

	stats := &Stats{
		NodesAtDepth: make([]int, maxDepth+1),
		PrunedAtDepth: make([]int, maxDepth+1),
		MaxDepth:maxDepth,
	}

	root := makeTree(bounds, mesh.Faces, 1, maxDepth, stats)
	voxels := getVoxels(root)

	// Save solutions .obj
	outputPath := filepath.Join("./test/solution", fmt.Sprintf("%s_voxel.obj", strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))))
	vertexCount, faceCount, err := writeOutput(voxels, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Terdapat error saat menyimpan solusi: %v\n", err)
		os.Exit(1)
	}

	// Print stats
	fmt.Printf("Voxels         : %d\n", len(voxels))
	fmt.Printf("Vertices       : %d\n", vertexCount)
	fmt.Printf("Faces          : %d\n", faceCount)
	
	fmt.Println("Nodes formed per depth:")
	for d := 1; d <= stats.MaxDepth; d++ {
		fmt.Printf("  %d : %d\n", d, stats.NodesAtDepth[d])
	}
	fmt.Println()

	fmt.Println("Nodes not traversed per depth:")
	for d := 1; d <= stats.MaxDepth; d++ {
		fmt.Printf("  %d : %d\n", d, stats.PrunedAtDepth[d])
	}
	fmt.Println()

	fmt.Printf("Max depth      : %d\n", stats.MaxDepth)
	fmt.Printf("Execution time : %v\n", time.Since(start))
	fmt.Printf("Output file    : %s\n", outputPath)

	runViewer(voxels, bounds)
}