ARG GOVERSION=1.22.2

FROM docker.io/library/golang:${GOVERSION}-alpine AS build
ARG GOVERSION
WORKDIR /
COPY ./ /

RUN go mod edit -C main -go $GOVERSION
RUN go mod tidy -C main

RUN go build -C main -o rebuilder main.go

FROM docker.io/library/golang:${GOVERSION}-alpine
ARG GOVERSION

COPY --from=build ./main/rebuilder /rebuilder

WORKDIR /
CMD [ "/rebuilder" ]

