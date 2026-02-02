FROM golang:alpine AS builder
WORKDIR /app
COPY . .
ARG GITHUB_SHA="unknown"
ARG VERSION="dev"
RUN echo "Building version: ${VERSION} commit: ${GITHUB_SHA:0:7}" && \
    go mod tidy && \
    go build -ldflags="-s -w -X main.Version=${VERSION} -X main.CurrentCommit=${GITHUB_SHA:0:7}" -trimpath -o subs-check .

FROM alpine
ENV TZ=Asia/Shanghai
RUN apk add --no-cache alpine-conf ca-certificates &&\
    /usr/sbin/setup-timezone -z Asia/Shanghai && \
    apk del alpine-conf && \
    rm -rf /var/cache/apk/*
COPY --from=builder /app/subs-check /app/subs-check
CMD ["/app/subs-check"]
EXPOSE 8199