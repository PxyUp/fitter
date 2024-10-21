FROM --platform=linux/arm64 arm64v8/golang:1.23

RUN go install github.com/playwright-community/playwright-go/cmd/playwright@latest \
      && playwright install --with-deps

WORKDIR /go/src/fitter_cli

COPY . .

RUN go mod download
RUN env GOOS=linux GOARCH=arm64  go build -o fitter_cli cmd/cli/main.go

CMD ["/fitter_cli"]