#!/bin/bash

function run() {
    trojan-go -config /etc/trojan-go/config.json
}

function ping_db() {
    while true; do mysql -hmysql -utrojan -ptrojan <<< "show databases" &> /dev/null && return; done;
}

function query_users() {
    mysql -hmysql -utrojan -ptrojan -Dtrojan -sN <<< "select * from users;"
}

function set_bandwidth() {
    while read -r _ _ PASSWORD _ _ _; do
        trojan-go -api-addr trojan:8080 -api set -add-profile \
            -target-password "$PASSWORD" && \
        trojan-go -api-addr trojan:8080 -api set -modify-profile \
            -target-password "$PASSWORD" \
            -ip-limit 5 \
            -upload-speed-limit 5242880 \
            -download-speed-limit 5242880
    done
}

function update_users() {
    while true; do query_users | set_bandwidth && sleep 5; done
}

ping_db

run &
sleep 2

update_users &

wait
