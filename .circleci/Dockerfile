FROM golang:1.12.6

WORKDIR /tusk

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh \
      | bash -s -- -b $GOPATH/bin v1.17.1
RUN go get github.com/jstemmer/go-junit-report
RUN apt-get update && \
      apt-get -y install rpm && \
      curl -sL https://git.io/goreleaser > /tmp/goreleaser && \
      chmod +x /tmp/goreleaser && \
      mv /tmp/goreleaser /usr/local/bin
RUN apt-get update && \
      apt-get install -y python-pip && \
      pip install --no-cache-dir \
      mkdocs==1.0.4 \
      mkdocs-rtd-dropdown==1.0.2 \
      Pygments==2.3.0

CMD [ "bash" ]
