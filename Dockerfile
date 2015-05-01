FROM google/golang

RUN go get github.com/tools/godep

RUN mkdir -p /gopath/src/github.com/lavab/lavabot
ADD . /gopath/src/github.com/lavab/lavabot
RUN cd /gopath/src/github.com/lavab/lavabot && godep go install

CMD []
ENTRYPOINT ["/gopath/bin/lavabot"]