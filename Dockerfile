FROM golang:1.4.2
MAINTAINER YI-HUNG JEN <yihungjen@macrodatalab.com>

COPY browser-bosh/bootstrapsrc/dist/css/bootstrap.min.css /static/
COPY browser-bosh/bosh/* /static/
COPY browser-bosh/c3src/c3.min.js /static/
COPY browser-bosh/c3src/c3.min.css /static/
COPY browser-bosh/d3src/d3.min.js /static/
COPY browser-bosh/handsontablesrc/dist/handsontable.full.js /static/
COPY browser-bosh/handsontablesrc/dist/handsontable.full.css /static/
COPY browser-bosh/jsonviewsrc/dist/jquery.jsonview.css /static/
COPY browser-bosh/jsonviewsrc/dist/jquery.jsonview.js /static/
COPY browser-bosh/papaparsesrc/papaparse.js /static/
COPY browser-bosh/urlsrc/url.min.js /static/
COPY browser-bosh/underscoresrc/underscore-min.js /static/
COPY browser-bosh/terminalsrc/src/* /static/

RUN mv /static/index.html /static/root.html

COPY alert.html /static/index.html

COPY main.go /go/src/github.com/macrodatalab/try-bigobject/
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
