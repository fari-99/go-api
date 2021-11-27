FROM golang:1.15

ARG APP_NAME
ENV APP_NAME=go-api

RUN mkdir -p /go/src/$APP_NAME

COPY .. /go/src/$APP_NAME

RUN apt-get update && apt-get install -y --no-install-recommends \
	git \
	&& rm -rf /var/lib/apt/lists/* \
	&& cd /go/src/$APP_NAME \
	&& rm -rf ./docker \
	&& rm -rf ./.git \
	&& go mod download \
    && go build cmd/servers/main/main.go

EXPOSE 8080

CMD ["/go/src/go-api/main"]
