FROM golang:alpine

WORKDIR /usr/src/app

EXPOSE 8088

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./

CMD ["app"]
