#!/bin/sh
/entrypoint.sh
mysql trojan < /initdb.sql
