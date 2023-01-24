FROM golang:bullseye AS build
COPY . /go/src/api
RUN cd src/api \
    && CGO_ENABLED=0 go build -o /go/bin/api main.go

FROM debian:bullseye
ENV PSQLVer 14
ENV POSTGRES_HOST 0.0.0.0
ENV POSTGRES_USER forum
ENV POSTGRES_PASSWORD 1
ENV POSTGRES_DB technopark

COPY --from=build /go/bin/api ./
RUN chmod +x api

RUN apt-get update 

RUN apt-get install -y tzdata
RUN ln -snf /usr/share/zoneinfo/Russia/Moscow /etc/localtime && echo Russia/Moscow > /etc/timezone

RUN apt-get install -y wget gnupg
RUN sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt bullseye-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN apt-get update && apt-get install postgresql-$PSQLVer -y
RUN chmod -R u=rwx /var/lib/postgresql/$PSQLVer/main/
RUN chmod -R 0700 /etc/postgresql/$PSQLVer/main

USER postgres
RUN /etc/init.d/postgresql start \
    && psql --command "CREATE USER $POSTGRES_USER WITH SUPERUSER PASSWORD '$POSTGRES_PASSWORD';" \
    && createdb -O $POSTGRES_USER $POSTGRES_DB\
    && /etc/init.d/postgresql stop

RUN echo "host all  all 0.0.0.0/0  md5" >> /etc/postgresql/$PSQLVer/main/pg_hba.conf
RUN echo "listen_addresses='*'\nsynchronous_commit = off\nfsync = off\nshared_buffers = 256MB\neffective_cache_size = 1536MB\n" >> /etc/postgresql/$PSQLVer/main/postgresql.conf
RUN echo "wal_buffers = 16MB\nmax_wal_size = 2GB\nrandom_page_cost = 1.0\nmax_connections = 100\nwork_mem = 8MB\nmaintenance_work_mem = 128MB" >> /etc/postgresql/$PSQLVer/main/postgresql.conf
RUN echo "full_page_writes = off" >> /etc/postgresql/$PSQLVer/main/postgresql.conf

COPY db db

CMD service postgresql start && ./api