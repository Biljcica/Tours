FROM golang:1.25-alpine AS builder

WORKDIR /app

# Kopiraj ceo Tours mikroservis
COPY ./Tours /app

# Kopiraj zajednički modul common u Tours folder
COPY ./Docker/common /app/common

WORKDIR /app
RUN go mod tidy

EXPOSE 8084
ENTRYPOINT ["go", "run", "main.go"]
