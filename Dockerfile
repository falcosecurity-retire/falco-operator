FROM sysdig/falco:0.12.1

USER root

RUN curl -L https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64 -o /usr/local/bin/dumb-init && \
    chmod +x /usr/local/bin/dumb-init

RUN mkdir -p /opt/bin/ \
  /var/falco-operator/rules

ADD build/falco-operator-amd64 /opt/bin/falco-operator

ENV PATH=${PATH}:/opt/bin

ENTRYPOINT ["dumb-init", "--", "falco-operator", "--", "/docker-entrypoint.sh"]

CMD []
