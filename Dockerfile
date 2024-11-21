FROM golang:1.23.1

# tetapin working directory nya
WORKDIR /app

# Copy semua file dari direktori lokal ke container
COPY . .

RUN go mod tidy

# Build aplikasi
RUN go build -o main .

# Perintah default untuk menjalankan aplikasi
CMD ["./main"]