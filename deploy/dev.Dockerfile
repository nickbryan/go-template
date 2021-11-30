FROM golang:1.16

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y \
    curl git make openssh-client && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN curl -sSL https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate.linux-amd64 /bin/migrate

RUN curl -sSfLo install.sh https://raw.githubusercontent.com/cosmtrek/air/master/install.sh \
    && chmod +x install.sh && sh install.sh && cp ./bin/air /bin/air

CMD air
