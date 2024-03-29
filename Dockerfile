FROM docker.servicewall.cn/alpine

WORKDIR /app

COPY ./porter /app/porter

CMD ["/app/porter"]