FROM golang:1.22

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest
RUN go install github.com/a-h/templ/cmd/templ@v0.2.648

COPY go.mod go.sum ./
RUN go mod download

CMD ["air"]