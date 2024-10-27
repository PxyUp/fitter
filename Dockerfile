FROM --platform=linux/arm64 arm64v8/golang:1.23

ENV GOARCH=arm64
ENV GOOS=linux

RUN go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps

WORKDIR /go/src/fitter_cli

ARG FITTER_CLI_VERSION

RUN wget -O fitter_cli https://github.com/PxyUp/fitter/releases/download/${FITTER_CLI_VERSION}/fitter_cli_${FITTER_CLI_VERSION}-linux-arm64

RUN chmod u+x fitter_cli

ENTRYPOINT ["/go/src/fitter_cli/fitter_cli"]