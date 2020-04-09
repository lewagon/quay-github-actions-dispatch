FROM alpine

ARG CRT=/certs/fullchain.pem
ARG KEY=/certs/privkey.pem
ARG GH_TOKEN=fake-token

ENV CRT $CRT
ENV KEY $KEY
ENV GH_TOKEN $GH_TOKEN
ENV DEBUG false

WORKDIR /app
COPY quay-github-actions-dispatch .

CMD ["./quay-github-actions-dispatch"]
