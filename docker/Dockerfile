FROM gcr.io/distroless/static-debian12 AS base

FROM base AS http-server

COPY ./bin/http-server /

CMD ["/http-server"]
