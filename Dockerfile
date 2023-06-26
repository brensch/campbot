# Base image
FROM golang:1.20

# Set the working directory in the container
WORKDIR /app

# Copy the application code into the container
COPY . .

# Build the application
RUN go build -o app

# Set the command to run the application
CMD ["./app"]