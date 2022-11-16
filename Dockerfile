FROM golang:alpine

WORKDIR /app

EXPOSE 8081

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
