FROM alpine:3.19

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file

VOLUME /pelldvs
WORKDIR /pelldvs
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["node", "--proxy_app", "kvstore"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh
COPY config-template.toml /etc/pelldvs/config-template.toml
