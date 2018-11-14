FROM golang:1.11.1

WORKDIR /tusk

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh \
      | bash -s -- -b $GOPATH/bin v1.12.2
RUN go get github.com/jstemmer/go-junit-report
RUN apt-get update && \
      apt-get -y install rpm && \
      curl -sL https://git.io/goreleaser > /tmp/goreleaser && \
      chmod +x /tmp/goreleaser && \
      mv /tmp/goreleaser /usr/local/bin
RUN apt-get update && \
      apt-get install -y python-pip && \
      pip install --no-cache-dir \
      mkdocs==0.17.3 \
      mkdocs-cinder==0.11.0 \
      Pygments==2.2.0

CMD [ "bash" ]
