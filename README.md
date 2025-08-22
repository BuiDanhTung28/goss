
<img width="400" height="400" alt="Go binding for faiss icon" src="https://github.com/user-attachments/assets/e8865cb8-cfd5-419c-8c35-a81c05f46cad" />

# Goss: New Production-Ready Go Binding for FAISS

Go bindings for [Faiss](https://github.com/facebookresearch/faiss), a library for efficient similarity search and clustering of dense vectors.

This library bundles the Faiss C++ source code and provides an automated build process, allowing you to use Faiss in your Go projects with a simple `go get` and `go generate` flow, without needing to manually build or install the Faiss library.

## Installation

This library requires a C++ compiler and build tools on your system to compile the bundled Faiss source code.

### Using prebuilt library

**For most users, you can simply install without building:**

```bash
go get github.com/BuiDanhTung28/goss
```

The library comes with prebuilt static libraries for common platforms:
- **macOS**: ARM64 (Apple Silicon) and x64 (Intel)
- **Linux**: x64 and ARM64

**Note**: If you encounter linking errors, you may need to build from source (see below).

### Building from source

If you need to build from source or the prebuilt libraries don't work for your platform:

### 1. Install Dependencies

First, install the required build tools for your operating system.

#### macOS

You will need **Xcode Command Line Tools** and **Homebrew**.

1.  **Install Xcode Command Line Tools:**
    ```sh
    xcode-select --install
    ```

2.  **Install Homebrew** (if you don't have it):
    Follow the instructions at [brew.sh](https://brew.sh/).

3.  **Install Dependencies with Homebrew:**
    ```sh
    brew install cmake libomp
    ```

#### Linux (Debian / Ubuntu)

```sh
sudo apt-get update && sudo apt-get install build-essential cmake
```

#### Linux (Fedora / CentOS / RHEL)

```sh
sudo yum groupinstall 'Development Tools' && sudo yum install cmake
```

### 2. Get and Build the Module

Once the dependencies are installed, you can get and build the module from within your project's directory.

1.  **Clone the Goss repository:**
    ```sh
    git clone https://github.com/BuiDanhTung28/goss.git
    cd goss
    ```

2.  **Build the Faiss static library:**
    This command compiles the C++ source code. It only needs to be run once per environment.
    ```sh
    go generate .
    ```
    *(This may take a few minutes the first time you run it.)*

3.  **In your project, use go get with replace:**
    ```sh
    # Get the module
    go get github.com/BuiDanhTung28/goss
    
    # Add replace directive to use local version
    go mod edit -replace github.com/BuiDanhTung28/goss=../goss
    ```
    
    **Or manually add to go.mod:**
    ```go
    module yourproject
    
    go 1.21
    
    require github.com/BuiDanhTung28/goss v0.0.0
    
    replace github.com/BuiDanhTung28/goss => ../goss
    ```

That's it! The library is now ready to be used in your project.

### Directory Structure
```
your-project/
├── go.mod
├── main.go
└── goss/           # Cloned Goss repo
    ├── internal/
    │   └── lib/
    │       ├── darwin_arm64/
    │       └── linux_x64/
    └── ...
```

## 🔧 API Reference

### Core Types

#### Index Interface
```go
type Index interface {
    D() int                    // Vector dimension
    IsTrained() bool          // Whether index is trained
    Ntotal() int64            // Number of indexed vectors
    MetricType() int          // Distance metric type
    Train(x []float32) error  // Train the index
    Add(x []float32) error    // Add vectors
    AddWithIDs(x []float32, xids []int64) error
    Search(x []float32, k int64) ([]float32, []int64, error)
    SearchBatch(queries []float32, k int64, batchSize int) ([][]float32, [][]int64, error)
    AddBatch(vectors []float32, batchSize int) error
    Reset() error             // Remove all vectors
    RemoveIDs(sel *IDSelector) (int, error)
    Delete()                  // Free memory
}
```

#### Index Types
```go
// Flat Index - Exact search, stores all vectors
type IndexFlat struct { Index }

// IVF Index - Inverted file with clustering
type IndexIVFFlat struct { Index }

// HNSW Index - Hierarchical Navigable Small World
type IndexHNSW struct { Index }
```

### Distance Metrics
```go
const (
    MetricInnerProduct  = 0 // Inner product (cosine for normalized vectors)
    MetricL2            = 1 // L2 (Euclidean) distance
    MetricL1            = 2 // L1 (Manhattan) distance
    MetricLinf          = 3 // L-infinity distance
    MetricLp            = 4 // Lp distance
    MetricCanberra      = 5 // Canberra distance
    MetricBrayCurtis    = 6 // Bray-Curtis distance
    MetricJensenShannon = 7 // Jensen-Shannon divergence
)
```

### Factory Functions
```go
// Create index from description string
func IndexFactory(d int, description string, metric int) (Index, error)

// Create specific index types
func NewIndexFlat(d int, metric int) (*IndexFlat, error)
func NewIndexFlatIP(d int) (*IndexFlat, error)      // Inner product
func NewIndexFlatL2(d int) (*IndexFlat, error)      // L2 distance
func NewIndexFlatL1(d int) (*IndexFlat, error)      // L1 distance
func NewIndexFlatLinf(d int) (*IndexFlat, error)    // L-infinity distance
```

## 📚 Usage Examples

### 1. Basic Flat Index (Exact Search)
```go
// Create index
index, err := faiss.NewIndexFlat(128, faiss.MetricL2)
defer index.Delete()

// Add vectors
vectors := make([]float32, 128*1000)
err = index.Add(vectors)

// Search
query := make([]float32, 128)
distances, labels, err := index.Search(query, 5)
```

### 2. IVF Index with Clustering
```go
// Create IVF index with 100 clusters
index, err := faiss.IndexFactory(128, "IVF100,Flat", faiss.MetricL2)
defer index.Delete()

// Train first, then add vectors
err = index.Train(trainingVectors)
err = index.Add(vectors)

// Search
distances, labels, err := index.Search(query, 10)
```

### 3. HNSW Index (Approximate Search)
```go
// Create HNSW index with 32 connections per node
index, err := faiss.IndexFactory(128, "HNSW32", faiss.MetricL2)
defer index.Delete()

// No training needed for HNSW
err = index.Add(vectors)

// Search
distances, labels, err := index.Search(query, 10)
```

### 4. Batch Operations
```go
// Add vectors in batches
err = index.AddBatch(vectors, 1000)

// Search multiple queries in batches
distances, labels, err := index.SearchBatch(queries, 10, 100)
```

### 5. Custom IDs and Vector Management
```go
// Add vectors with custom IDs
customIDs := make([]int64, numVectors)
err = index.AddWithIDs(vectors, customIDs)

// Remove specific vectors
selector := faiss.NewIDSelectorRange(100, 200)
removed, err := index.RemoveIDs(selector)
```

### 6. Index I/O Operations
```go
// Save index to file
err = faiss.WriteIndex(index, "my_index.faiss")

// Load index from file
loadedIndex, err := faiss.ReadIndex("my_index.faiss")
defer loadedIndex.Delete()
```

## 🎯 Index Type Recommendations

- **Flat Index**: Small datasets (< 100K vectors), exact search needed
- **IVF Index**: Large datasets (100K - 10M vectors), good accuracy needed
- **HNSW Index**: Large datasets, very fast search needed, approximate search OK
- **PQ Index**: Very large datasets (> 10M vectors), memory constrained

## ⚡ Performance Tips

1. **Choose the right index type** for your use case
2. **Use batch operations** for large datasets
3. **Train IVF indexes** with representative data
4. **Use appropriate batch sizes** (100-1000 vectors per batch)
