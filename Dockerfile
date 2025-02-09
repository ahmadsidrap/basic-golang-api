FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY bin /app
COPY .env /app/

# Expose port if your app listens on one
EXPOSE 8080

# Command to run the executable
CMD ["/app/basic-golang-api"]
