# --- GIAI ĐOẠN 1: MÔI TRƯỜNG BUILD ---
    FROM ubuntu:latest AS builder

    # Cài đặt các dependencies cần thiết cho Faiss
    RUN apt-get update && apt-get install -y \
        build-essential \
        cmake \
        libopenblas-dev \
        liblapack-dev \
        git \
        libgflags-dev
    
    # Đặt thư mục làm việc trong container
    WORKDIR /faiss-build
    
    # Copy mã nguồn Faiss vào container
    COPY faiss_source ./faiss_source
    
    # Cấu hình với CMake
    RUN cmake -S faiss_source -B build \
        -DFAISS_ENABLE_GPU=OFF \
        -DFAISS_ENABLE_PYTHON=OFF \
        -DBUILD_SHARED_LIBS=OFF \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_POSITION_INDEPENDENT_CODE=ON
    
    # Biên dịch Faiss
    RUN cmake --build build --target faiss --config Release
    
    # --- GIAI ĐOẠN 2: ĐÓNG GÓI THƯ VIỆN CUỐI CÙNG ---
    # Sử dụng một image siêu nhỏ để chứa file thư viện cuối cùng.
    FROM scratch
    
    # Copy file thư viện đã build từ giai đoạn builder
    COPY --from=builder /faiss-build/build/faiss/libfaiss.a /libfaiss.a
    
    # Cung cấp một lệnh mặc định, mặc dù không cần thiết
    # nhưng giúp tránh lỗi "no command specified"
    CMD ["/bin/sh"]