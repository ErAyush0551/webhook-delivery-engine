# ---------- Stage 1: Build ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Pehle sirf go.mod/go.sum copy karo — isse Docker layer caching kaam karega
COPY go.mod go.sum ./
RUN go mod download

# Ab poora source code copy karo
COPY . .

# Static binary banao (CGO_ENABLED=0 zaroori hai alpine ke liye)
RUN CGO_ENABLED=0 GOOS=linux go build -o hookflow-server main.go

# ---------- Stage 2: Run ----------
FROM alpine:3.20

WORKDIR /app

# Migration files bhi chahiye honge runtime pe
COPY --from=builder /app/migration ./migration
COPY --from=builder /app/hookflow-server .

EXPOSE 8080

CMD ["./hookflow-server"]

#we made this because we want to use multi-stage builds to reduce the final image size. The first stage builds the Go application in a Golang environment, while the second stage uses a lightweight Alpine image to run the compiled binary. 
#This approach ensures that we only include the necessary files in the final image, making it smaller and more efficient.