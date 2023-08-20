# BUILD STAGE
FROM golang:1.21-alpine as builder
ARG GITHUB_SSH_KEY
RUN apk add --no-cache git openssh && \
    mkdir /root/.ssh && \
    echo "$GITHUB_SSH_KEY" > /root/.ssh/id_rsa && \
    chmod 400 /root/.ssh/id_rsa && \
    touch /root/.ssh/known_hosts && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts && \
    git config --global url.git@github.com:.insteadOf https://github.com/
WORKDIR /repo
COPY . .
RUN go build -o app .

# FINAL STAGE
FROM alpine:3.18
EXPOSE 8000
RUN apk add --no-cache ca-certificates
COPY --from=builder /repo/app /usr/local/bin/
RUN chown -R nobody:nogroup /usr/local/bin/app
RUN deluser --remove-home root || echo "DONE"
USER nobody
CMD [ "app" ]
