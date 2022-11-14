FROM golang:1.19 as development

WORKDIR /app/wishlist_bot

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/cespare/reflex@latest

EXPOSE 4000

CMD reflex -r '\.go$' go run app/main.go --start-service