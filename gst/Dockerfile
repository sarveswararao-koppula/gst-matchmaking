# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.13

# Set the working directory to the project directory
WORKDIR /go/src/mm

# Set the timezone to IST (Indian Standard Time)
ENV TZ=Asia/Kolkata

# Copy the local package files to the container's workspace.
COPY . .

# Copy Go dependencies from the local vendor directory
RUN mkdir -p /go/src/github.com && cp -r /go/src/mm/vendor/github.com/* /go/src/github.com/

#module turn off
#ENV GO111MODULE=off

# Build the project
RUN go build -o /go/bin/mm

# Run the mm binary.
ENTRYPOINT ["/go/bin/mm"]

# Document that the service listens on port 3000.
EXPOSE 8080
