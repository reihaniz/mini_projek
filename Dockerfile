# Menggunakan base image Golang
FROM golang:1.23.1

# Set working directory di dalam container
WORKDIR /app

# Copy semua file dari direktori lokal ke dalam container
COPY . .

# Download dependencies (jika menggunakan go modules)
RUN go mod tidy

# Build aplikasi
RUN go build -o main .

# Perintah default untuk menjalankan aplikasi
CMD ["./main"]