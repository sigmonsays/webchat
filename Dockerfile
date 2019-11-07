

FROM golang:1.11-stretch AS build

WORKDIR /go/src/github.com/sigmonsays/webchat
RUN apt-get install \
    git gcc g++ binutils
COPY . /go/src/github.com/sigmonsays/webchat/
RUN go get -d .
ENV GOPATH=/go
RUN go install -ldflags '-w -extldflags "-static"' github.com/sigmonsays/webchat/...

# ---

FROM golang:1.11-stretch
COPY --from=build /go/bin/chat /webchat
COPY --from=build /go/src/github.com/sigmonsays/webchat/static /go/static
EXPOSE  8080
CMD [ "/webchat", "-static", "/go/static"  ]
