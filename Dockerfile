FROM golang:1.24
WORKDIR /SocketServer
COPY . .
RUN go mod download
EXPOSE 8082 8081
CMD [ "go","run","." ]