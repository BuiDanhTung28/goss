
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