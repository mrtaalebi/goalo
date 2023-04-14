#!/bin/bash

function run() {
    trojan-go -config /etc/trojan-go/config.json
}

function ping_db() {
    while true; do mysql -h mysql -u trojan -p trojan -e 'show databases' && return; done;
}

function get_users() {
    mysql -h mysql -u trojan -p trojan <<< """
        \u trojan
        select * from users;
    """
}

function set_bandwidth() {
    while IFS=$'\n' read -r _ _ PASSWORD _ _ _; do
        trojan-go -api-addr trojan:8080 -api set -modify-profile \
            -target-password "$PASSWORD" \
            -ip-limit 5 \
            -upload-speed-limit 5242880 \
            -download-speed-limit 5242880 \
            -quota 0
    done
}

run &
sleep 2

ping_db
get_users
set_bandwidth

fg

