FROM golang:1.23.2-alpine3.20 as build
WORKDIR /go/src/betting

RUN go install github.com/air-verse/air@latest
COPY go.* ./
RUN go mod download
COPY . .

RUN go build -o /go/bin/betting ./cmd/betting

FROM build as development

WORKDIR /go/src/betting

CMD ["air", "-c", "./build/betting/.air.toml"]

