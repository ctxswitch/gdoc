#FROM golang:alpine3.14 AS BUILDER
FROM golang:1.17 AS builder
ARG VERSION
ARG BUILD

ENV GODOC_ROOT=/usr/local/go

WORKDIR /usr/src/godoc-web
COPY go.mod go.sum ./
RUN : \
  && go mod download && go mod verify \
  && go install -v golang.org/x/tools/cmd/godoc@latest \
  && :

COPY . .
RUN : \
  && go build -ldflags "-X=main.Version=${VERSION} -X=main.Build=${BUILD}" -o /usr/local/bin ./... \
  && :

WORKDIR /
RUN rm -rf /usr/src/godoc-web

EXPOSE 6060
CMD [ "/usr/local/bin/godoc-web" ]
