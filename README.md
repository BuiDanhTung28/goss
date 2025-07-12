# FAISS Go Binding

Go binding cho Facebook AI Similarity Search (FAISS) library sử dụng CGO.

## Yêu cầu

1. **FAISS C API**: Bạn cần build và cài đặt FAISS với C API support:

```bash
git clone https://github.com/facebookresearch/faiss.git
cd faiss
cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON .
make -C build
sudo make -C build install
sudo cp build/c_api/libfaiss_c.so /usr/lib
```

2. **Go 1.21+**: Đảm bảo bạn có Go version 1.21 trở lên.

3. **GCC/Clang**: Cần có C compiler để build CGO.

## Cài đặt

```bash
go mod tidy
```

## Sử dụng

### Tạo Index

```go
package main

import (
    "log"
    "faiss-go"
)

func main() {
    // Tạo index với dimension 128, loại IndexFlatL2
    index, err := faiss.NewIndex(128, faiss.IndexFlatL2)
    if err != nil {
        log.Fatal(err)
    }
    defer index.Free()
}
```

### Thêm Vectors

```go
// Tạo vectors (1000 vectors, mỗi vector 128 chiều)
vectors := make([]float32, 1000*128)
// ... khởi tạo dữ liệu vectors ...

// Thêm vào index
err = index.Add(vectors)
if err != nil {
    log.Fatal(err)
}
```

### Tìm kiếm

```go
// Tạo query vector
query := make([]float32, 128)
// ... khởi tạo query ...

// Tìm 5 vectors gần nhất
labels, distances, err := index.Search(query, 5)
if err != nil {
    log.Fatal(err)
}

// In kết quả
for i, label := range labels {
    fmt.Printf("ID: %d, Distance: %.4f\n", label, distances[i])
}
```

### Batch Search

```go
// Tạo nhiều query vectors
queries := make([]float32, 10*128) // 10 queries, mỗi query 128 chiều
// ... khởi tạo queries ...

// Batch search
labels, distances, err := index.SearchBatch(queries, 5)
if err != nil {
    log.Fatal(err)
}
```

### Lưu và Load Index

```go
// Lưu index
err = index.SaveIndex("my_index.faiss")
if err != nil {
    log.Fatal(err)
}

// Load index
loadedIndex, err := faiss.LoadIndex("my_index.faiss")
if err != nil {
    log.Fatal(err)
}
defer loadedIndex.Free()
```

## Các loại Index được hỗ trợ

- `IndexFlatL2`: Exact search với L2 distance
- `IndexFlatIP`: Exact search với Inner Product
- `IndexIVFFlat`: Inverted File với Flat quantizer
- `IndexIVFPQ`: Inverted File với Product Quantizer

## Chạy Example

```bash
go run example/main.go
```

## API Reference

### Index

```go
type Index struct {
    ptr unsafe.Pointer
}
```

### Methods

- `NewIndex(dimension int, indexType IndexType) (*Index, error)`: Tạo index mới
- `Add(vectors []float32) error`: Thêm vectors vào index
- `Search(query []float32, k int) ([]int64, []float32, error)`: Tìm kiếm single query
- `SearchBatch(queries []float32, k int) ([]int64, []float32, error)`: Batch search
- `Train(vectors []float32) error`: Train index (cho các index cần training)
- `GetDimension() int`: Lấy dimension của index
- `GetTotal() int`: Lấy tổng số vectors trong index
- `IsTrained() bool`: Kiểm tra index đã được train chưa
- `Reset()`: Xóa tất cả vectors
- `Free()`: Giải phóng memory
- `SaveIndex(filename string) error`: Lưu index ra file
- `LoadIndex(filename string) (*Index, error)`: Load index từ file

## Troubleshooting

### Lỗi "faiss_c.h not found"

Đảm bảo FAISS C API đã được cài đặt đúng cách và header files có trong `/usr/local/include`.

### Lỗi "libfaiss_c.so not found"

Đảm bảo shared library đã được copy vào `/usr/lib` hoặc thêm đường dẫn vào `LD_LIBRARY_PATH`.

### Lỗi CGO

Đảm bảo có C compiler (GCC hoặc Clang) và các development tools cần thiết.

## License

MIT License 