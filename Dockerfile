FROM golang:latest

ADD . /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENTRYPOINT ["go","run","."]