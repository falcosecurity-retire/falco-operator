FROM alpine:3.6

RUN adduser -D falco-operator
USER falco-operator

ADD tmp/_output/bin/falco-operator /usr/local/bin/falco-operator
