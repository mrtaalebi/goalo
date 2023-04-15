FROM teddysun/trojan-go:0.10.6
ENV TZ Asia/Tehran
RUN apk add bash mysql-client coreutils --no-cache
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
