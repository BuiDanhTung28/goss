# Go-Faiss: Go bindings for Faiss

Go bindings for [Faiss](https://github.com/facebookresearch/faiss), a library for efficient similarity search and clustering of dense vectors.

This library bundles the Faiss C++ source code and provides an automated build process, allowing you to use Faiss in your Go projects with a simple `go get` and `go generate` flow, without needing to manually build or install the Faiss library.

## Features

-   **Self-contained:** Bundles Faiss source code, no need for manual installation.
-   **Automated Build:** Uses `go generate` to automatically compile the required static library.
-   **Cross-Platform:** Provides build scripts and instructions for macOS and Linux.
-   **Memory Safe:** Uses Go finalizers to manage the lifecycle of C objects.
-   **Idiomatic Go API:** Provides a Go-friendly API on top of the Faiss C API.

## Installation

This library requires a C++ compiler and build tools on your system to compile the bundled Faiss source code.

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

1.  **Get the module:**
    *(Assuming your project is already a Go module. Replace with your actual repo path if needed.)*
    ```sh
    go get -d ./...
    ```

2.  **Build the Faiss static library:**
    This command compiles the C++ source code. It only needs to be run once per environment.
    ```sh
    go generate .
    ```
    *(This may take a few minutes the first time you run it.)*

That's it! The library is now ready to be used in your project.

## Usage

Here is a simple example of how to create an index, add vectors, and perform a search.

```go
package main

import (
	"fmt"
	"log"

	"goss/faiss" // Use your module name
)

func main() {
	dimension := 8

	// Create a Flat L2 index. 
	// Note: NewIndexFlatL2 is not a real function yet. 
	// You would use NewIndexFlat(dimension, faiss.MetricL2)
	idx, err := faiss.NewIndexFlat(dimension, faiss.MetricL2)
	if err != nil {
		log.Fatal(err)
	}
	defer idx.Delete()

	fmt.Printf("Index created. Is trained: %v, D: %d, Ntotal: %d\n", idx.IsTrained(), idx.D(), idx.Ntotal())

	// Some vectors to add to the index
	vectors := []float32{
		1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0,
		2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0,
		3.0, 3.0, 3.0, 3.0, 3.0, 3.0, 3.0, 3.0,
	}

	// Add vectors to the index
	err = idx.Add(vectors)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("After adding, Ntotal: %d\n", idx.Ntotal())

	// A query vector
	query := []float32{1.1, 1.1, 1.1, 1.1, 1.1, 1.1, 1.1, 1.1}
	k := int64(2) // Number of nearest neighbors to search for

	// Search the index
	distances, labels, err := idx.Search(query, k)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Search results for k=%d:\n", k)
	for i := 0; i < int(k); i++ {
		fmt.Printf("  - Rank %d: ID=%d, Distance=%.4f\n", i+1, labels[i], distances[i])
	}
}
```

## How It Works

This library uses `cgo` to call the Faiss C API. To avoid forcing users to manually install the correct version of Faiss, we bundle the Faiss source code as a `git submodule`.

When you run `go generate`, it executes a build script (`build.sh`) that:
1.  Checks for the required system dependencies (`cmake`, `make`, etc.).
2.  Compiles the bundled C++ source code into a static library (`libfaiss.a`).
3.  Places the library inside an `internal` directory, which is ignored by Git.

When you run `go build`, platform-specific build tags (`//go:build`) ensure that Go uses the correct CGO flags to link your program against the pre-compiled `libfaiss.a`. This results in a self-contained, static binary. 