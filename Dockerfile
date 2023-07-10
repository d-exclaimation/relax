# --- Stage 1: Dependencies and building ---
FROM golang:alpine as builder

WORKDIR /app

# Install dependencies first
COPY go.mod .
RUN go mod download

# Copy the rest of the files and build
COPY . .
RUN go build -o main .

# --- Stage 2: Setting up runner image ---

FROM alpine as runner

WORKDIR /app

# Set mode to production
ENV GO_ENV=production

# Get the binary from the builder stage
COPY --from=builder /app/main .

CMD ["./main"]