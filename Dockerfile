FROM golang:1.4.2
MAINTAINER YI-HUNG JEN <yihungjen@macrodatalab.com>

COPY html_content/ /static/
COPY alert.html /static/

COPY main.go /go/src/github.com/macrodatalab/try-bigobject/
COPY docker.go /go/src/github.com/macrodatalab/try-bigobject/
COPY discovery/ /go/src/github.com/macrodatalab/try-bigobject/discovery/
COPY proxy/ /go/src/github.com/macrodatalab/try-bigobject/proxy/
COPY Godeps/ /go/src/github.com/macrodatalab/try-bigobject/Godeps/
WORKDIR /go/src/github.com/macrodatalab/try-bigobject

ENV GOPATH /go/src/github.com/macrodatalab/try-bigobject/Godeps/_workspace:$GOPATH
RUN go install

ENV TRIAL_SERVICE_ENDPOINT try.bigobjectio
ENV TRIAL_SERVICE_IMAGE macrodata/bigobject-dev

EXPOSE 80

ENTRYPOINT ["try-bigobject"]
