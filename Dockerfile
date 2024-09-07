# Build stage
FROM --platform=linux/$TARGETARCH golang:1.22.6-alpine as build

# Install dependencies
RUN apk --no-cache add git gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files (if available) to avoid re-downloading dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the project files
COPY . .

# Build the application
RUN go build -o cmd -tags musl --ldflags '-linkmode external -extldflags "-static -s"' .

# Final stage: build minimal image
FROM --platform=linux/$TARGETARCH gcr.io/distroless/static-debian12 as release

# Copy the compiled binary from the build stage
COPY --from=build /app/cmd /cmd

# Specify the command to run
CMD ["/cmd"]

