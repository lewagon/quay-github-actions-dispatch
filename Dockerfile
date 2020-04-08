FROM alpine

ARG CRT=/etc/letsencrypt/live/quay-bridge.lewagon.co/fullchain.pem
ARG KEY=/etc/letsencrypt/live/quay-bridge.lewagon.co/privkey.pem
ARG GH_TOKEN=fake-token

ENV CRT $CRT
ENV KEY $KEY
ENV GH_TOKEN $GH_TOKEN

WORKDIR /app
COPY quay-github-actions-dispatch .

# COPY . .
# RUN go mod download
# RUN go mod verify
# RUN go build -o quay-github-actions-dispatch

CMD ["./quay-github-actions-dispatch"]
