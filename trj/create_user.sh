#!/bin/bash
set -u

function generate_password() {
    cat /dev/urandom | base64 -w 0 | head -c 32
}

function insert_user() {
    mysql -hmysql -utrojan -ptrojan -Dtrojan -sN <<< "insert into users(username, rawpassword) values('$REMARKS', '$PASSWORD');"
}

function generate_connstr() {
    printf 'trojan://%s@shop.pandaha.work?security=tls&alpn=http/1.1&headerType=none&fp=chrome&type=tcp&sni=shop.pandaha.work#f%s' "$PASSWORD" "$REMARKS"
}

function main() {
    REMARKS=$1
    PASSWORD=$(generate_password)
    insert_user
    generate_connstr
}

main $1
