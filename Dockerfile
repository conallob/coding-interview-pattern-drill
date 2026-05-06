FROM --platform=linux/amd64 alpine:3 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY coding-interview-pattern-drill /usr/local/bin/coding-interview-pattern-drill
EXPOSE 7777
ENTRYPOINT ["coding-interview-pattern-drill"]
# Default to serve mode; pass LEETCODE_SESSION env var for credentials.
# Override with no args to use CLI mode (requires -it).
CMD ["serve", "--no-open"]
