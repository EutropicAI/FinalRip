FROM golang:1.24-bookworm AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /build

COPY . .

RUN go mod download

RUN make worker

FROM debian:bookworm AS app

# prepare environment
RUN apt update -y && apt upgrade -y && \
    apt install -y wget && \
    wget -O /etc/apt/keyrings/gpg-pub-moritzbunkus.gpg https://mkvtoolnix.download/gpg-pub-moritzbunkus.gpg && \
    apt update -y && apt upgrade -y

RUN apt install -y mkvtoolnix

WORKDIR /app

ENV TZ=Asia/Shanghai

ENV FINALRIP_EASYTIER_HOST 10.126.126.251

COPY --from=mwader/static-ffmpeg:8.0 /ffmpeg /usr/local/bin/
COPY --from=mwader/static-ffmpeg:8.0 /ffprobe /usr/local/bin/

COPY --from=builder /build/worker/worker /app/
COPY --from=builder /build/conf/finalrip.yml /app/conf/

CMD ["/app/worker", "merge"]
