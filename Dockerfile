FROM golang:1.11.2-alpine3.8

ARG LIBRESSL_VERSION=1.1
ARG LIBRDKAFKA_VERSION=0.11.6-r1

RUN apk add libcrypto${LIBRESSL_VERSION} libssl${LIBRESSL_VERSION} --update-cache --repository http://nl.alpinelinux.org/alpine/edge/main && \
	apk add make \
	  pkgconfig \
	  bash \
	  gcc \
	  libc-dev &&\
\
    apk add librdkafka=${LIBRDKAFKA_VERSION} librdkafka-dev=${LIBRDKAFKA_VERSION} --update-cache --repository http://nl.alpinelinux.org/alpine/edge/community

