# Golang image where workspace (GOPATH) configured at /go.
FROM golang:1.9

# Copy the local package files to the container’s workspace.
ADD . /go/src/github.org/leedark/fileLogger

# Setting up working directory
WORKDIR /go/src/github.org/leedark/fileLogger

# TODO: organize deps
# Get all dependencies
RUN go get -v ./...

# Just build app
RUN go build -v

# Command liine arguments
ARG app_address
ENV app_address ${app_address}
ARG app_port
ENV app_port ${app_port}
#RUN echo $app_port

#ENTRYPOINT ["./rest/rest_service"]
CMD ./fileLogger start --address ${app_address} --port ${app_port}

EXPOSE 5000