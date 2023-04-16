#!/bin/bash
set -u

function generate_password() {
    cat /dev/urandom | base64 -w 0 | head -c 32
}

function insert_user() {
    docker compose exec -it mysql sh -c """mysql -hmysql -utrojan -ptrojan -Dtrojan -sN <<< \"insert into users(username, rawpassword) values('$REMARKS', '$PASSWORD');\"""" &>/dev/null
}

function generate_connstr() {
    printf '\ntrojan://%s@shop.pandaha.work:443?security=tls&alpn=http/1.1&headerType=none&type=tcp&sni=shop.pandaha.work#%s\n' "$PASSWORD" "$REMARKS"
}

function main() {
    REMARKS=$1
    PASSWORD=$(generate_password)
    insert_user
    generate_connstr
}

main $1
