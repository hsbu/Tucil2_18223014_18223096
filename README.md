# 3D Voxelization with Octree

## Deskripsi Program
Program merupakan implementasi **voxelisasi model 3D** menggunakan struktur data **Octree** dengan bahasa Go. Program menerima file model 3D berformat `.obj` dan mengubahnya menjadi representasi voxel (kotak-kotak kecil 3D) melalui proses subdivisi ruang secara rekursif.

Proses voxelisasi dilakukan dengan langkah-langkah berikut:
1. Parsing file `.obj` untuk membaca vertices dan faces (triangle mesh)
2. Membentuk Axis-Aligned Bounding Box (AABB) yang membungkus seluruh mesh
3. Melakukan subdivisi ruang secara rekursif menggunakan Octree hingga kedalaman maksimum yang ditentukan
4. Pada setiap node, dilakukan pengecekan interseksi triangle-AABB menggunakan Separating Axis Theorem (SAT)
5. Node yang tidak berpotongan dengan mesh akan dipangkas (*pruned*)
6. Hasil voxelisasi disimpan sebagai file `.obj` baru dan divisualisasikan secara interaktif menggunakan 3D wireframe viewer berbasis Ebiten

## Requirements
- **Go 1.25.6** atau lebih tinggi
- **Ebiten v2** (dependency untuk 3D viewer, otomatis di-download via `go mod`)

## Instalasi
1. Pastikan Go sudah terinstal di sistem Anda. Cek dengan menjalankan:
   ```bash
   go version
   ```

2. Clone repositori ini:
   ```bash
   git clone https://github.com/hsbu/Tucil2_18223014_18223096.git
   cd Tucil2_18223014_18223096
   ```

3. Download dependencies:
   ```bash
   cd src
   go mod tidy
   ```

## Cara Mengkompilasi Program

### Kompilasi manual
```bash
cd src
go build -o ../bin/tucil2.exe .
```

File executable akan tersimpan di folder `bin/`.

## Cara Menjalankan Program
Jalankan program dari **root folder** proyek (`Tucil2_18223014_18223096/`).

### Run dengan file executable

#### Windows (PowerShell/Command Prompt)
```bash
.\bin\tucil2.exe <input.obj> <max_depth>
```

**Contoh:**
```bash
.\bin\tucil2.exe .\test\data\teddybear.obj 6
```

#### Linux/macOS
```bash
./bin/tucil2 <input.obj> <max_depth>
```

### Parameter
| Parameter | Deskripsi |
|-----------|-----------|
| `<input.obj>` | Path ke file model 3D berformat `.obj` |
| `<max_depth>` | Tingkat kedalaman maksimum Octree (bilangan bulat positif, semakin besar semakin detail) |

## Cara Menggunakan Program

1. Jalankan program dengan perintah di atas beserta parameter yang diinginkan
2. Program akan memproses file `.obj` dan menampilkan statistik:
   - Jumlah voxel yang dihasilkan
   - Jumlah vertices dan faces pada output
   - Jumlah node yang terbentuk dan dipangkas per kedalaman
   - Waktu eksekusi
   - Path file output

3. Setelah proses selesai, **3D Wireframe Viewer** akan terbuka secara otomatis untuk memvisualisasikan hasil voxelisasi:
   - **Klik kiri + drag** untuk merotasi model
   - **Scroll** untuk zoom in/out

4. File hasil voxelisasi tersimpan otomatis di folder `test/solution/` dengan format `<nama_input>_voxel.obj`

## Format File Input
File input berupa file **Wavefront OBJ** standar yang berisi triangle mesh. Contoh format:
```
v 0.0 1.0 0.0
v 1.0 0.0 0.0
v 0.0 0.0 1.0
f 1 2 3
```
- Baris `v` mendefinisikan vertex (x, y, z)
- Baris `f` mendefinisikan face berupa indeks 3 vertex (triangle)

File input contoh tersedia di folder `test/data/`:
- `cow.obj`
- `line.obj`
- `pumpkin.obj`
- `teapot.obj`
- `teddybear.obj`

## Struktur Folder
```
Tucil2_18223014_18223096/
├── bin/                        # File executable
│   └── tucil2.exe
├── doc/                        # Dokumentasi
├── src/                        # Source code
│   ├── main.go                 # Program utama (parsing, octree, voxelisasi)
│   ├── viewer.go               # 3D wireframe viewer (Ebiten)
│   ├── go.mod                  # Go module definition
│   └── go.sum                  # Go module checksums
├── test/                       # File testing
│   ├── data/                   # File input .obj contoh
│   │   ├── cow.obj
│   │   ├── line.obj
│   │   ├── pumpkin.obj
│   │   ├── teapot.obj
│   │   └── teddybear.obj
│   └── solution/               # File output hasil voxelisasi
│       ├── cow_voxel.obj
│       ├── line_voxel.obj
│       ├── pumpkin_voxel.obj
│       ├── teapot_voxel.obj
│       └── teddybear_voxel.obj
├── .gitignore
└── README.md                   # File ini
```

## Author
**Nama**: Muhamad Hasbullah Faris </br>
**NIM**: 18223014

**Nama**: Matthew Sebastian Kurniawan </br>
**NIM**: 18223096