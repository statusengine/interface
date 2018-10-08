#!/bin/bash

mysql -p"3306" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -D"${MYSQL_DATABASE}" < /tmp/statusengine.sql
