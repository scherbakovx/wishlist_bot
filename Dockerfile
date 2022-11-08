FROM golang:1.19 as development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 4000

CMD go run app/main.go