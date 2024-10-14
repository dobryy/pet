FROM golang:1.23.2-alpine3.20 as build
WORKDIR /go/src/mybets

RUN go install github.com/air-verse/air@latest
COPY go.* ./
RUN go mod download
COPY . .

RUN go build -o /go/bin/mybets ./cmd/mybets

FROM build as development

WORKDIR /go/src/mybets

CMD ["air", "-c", "./build/mybets/.air.toml"]

