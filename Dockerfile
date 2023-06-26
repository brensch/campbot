# Build stage
FROM golang:1.20 AS build-env

# Set the working directory in the build stage container
WORKDIR /src

# Copy the application code into the build stage container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM scratch

# Set the working directory in the final stage container
WORKDIR /app

# Copy the root certificates from the build stage container
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled application from the build stage container
COPY --from=build-env /src/app /app/

COPY campgrounds.json /app/

# Set the command to run the application
CMD ["./app"]