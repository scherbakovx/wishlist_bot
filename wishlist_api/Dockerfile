FROM golang:1.19 as development

WORKDIR /app/wishlist_api/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/cespare/reflex@latest

EXPOSE 3000

CMD reflex -r '\.go$' go run main.go --start-service