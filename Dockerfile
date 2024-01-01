FROM golang:1.20.5-alpine3.18 AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
# copy static files
COPY --from=builder /app/assets ./assets
# install ffmpeg
RUN apk add --no-cache ffmpeg
# install ytarchive
RUN wget https://github.com/Kethsar/ytarchive/releases/download/latest/ytarchive_linux_amd64.zip
RUN unzip ytarchive_linux_amd64.zip
RUN rm ytarchive_linux_amd64.zip
RUN chmod +x ytarchive
RUN mv ytarchive /usr/local/bin/

CMD ["./main"]
