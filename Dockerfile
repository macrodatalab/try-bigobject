FROM ubuntu:latest
MAINTAINER YI-HUNG JEN <yihungjen@macrodatalab.com>

COPY html_content/ /static/
COPY bin/try-bigobject /usr/local/bin/try-bigobject

ENV REGISTRY_URI http://localhost:5000

EXPOSE 80

CMD ["try-bigobject"]
