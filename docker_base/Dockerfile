FROM --platform=linux/arm64 arm64v8/golang:1.23

ENV GOARCH=arm64
ENV GOOS=linux

RUN go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps