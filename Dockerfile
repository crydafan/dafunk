FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/dafunk .

FROM alpine:3.20

RUN apk update && apk add --no-cache ffmpeg
RUN adduser -D appuser

WORKDIR /app

COPY --from=build /out/dafunk /app/dafunk
# TODO: Move this to a bucket instead of copying it into the image
COPY musique /app/musique 

USER appuser

EXPOSE 8080

CMD [ "/app/dafunk" ]