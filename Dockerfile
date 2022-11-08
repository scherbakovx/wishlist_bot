FROM golang:1.19 as development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/cespare/reflex@latest

EXPOSE 4000

CMD reflex -g '*.go' go run app/main.go --start-service