#!/bin/bash

function run() {
    trojan-go -config /etc/trojan-go/config.json
}

function ping_db() {
    while true; do mysql -hmysql -utrojan -ptrojan <<< "show databases" &> /dev/null && return; done;
}

function query_users() {
    mysql -hmysql -utrojan -ptrojan -Dtrojan -sN <<< "select rawpassword from users;"
}

function update_usage() {
    while read -r RAWPASSWORD; do
        HASH=$(echo -n "$RAWPASSWORD" | openssl dgst -sha224 | cut -d' ' -f2)
        mysql -hmysql -utrojan -ptrojan -Dtrojan <<< "update users set password='$HASH', quota=62914560, download=0, upload=0 where rawpassword='$RAWPASSWORD'"
        trojan-go -api-addr trojan:8080 -api set -add-profile \
            -target-password "$RAWPASSWORD" &>/dev/null && \
        trojan-go -api-addr trojan:8080 -api set -modify-profile \
            -target-hash "$HASH" \
            -ip-limit 1 &>/dev/null
    done
}

function update_users() {
    while true; do query_users | update_usage && sleep 60; done
}

ping_db

run &
sleep 2

update_users &

wait
