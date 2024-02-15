FROM golang:1.21.7 AS development

WORKDIR /app/dapex-be

COPY go.mod go.sum main.go ./

RUN go mod download

RUN go build -o bin .

EXPOSE 4000

CMD ["./bin"]