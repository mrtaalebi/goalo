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
        HASH=$(echo "$RAWPASSWORD" | sha224sum -z | cut -d' ' -f1)
        mysql -hmysql -utrojan -ptrojan -Dtrojan <<< "update users set password='$HASH', quota=10485760, download=0, upload=0 where rawpassword='$RAWPASSWORD'"
        trojan-go -api-addr trojan:8080 -api set -add-profile \
            -target-password "$RAWPASSWORD" &>/dev/null && \
        trojan-go -api-addr trojan:8080 -api set -modify-profile \
            -target-password "$RAWPASSWORD" \
            -ip-limit 5 \
            -upload-speed-limit 2097152 \
            -download-speed-limit 2097152 &>/dev/null
    done
}

function update_users() {
    while true; do query_users | update_usage && sleep 5; done
}

ping_db

run &
sleep 2

update_users &

wait
