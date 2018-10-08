FROM php:7.2-fpm

COPY . /usr/share/statusengine-ui
WORKDIR /usr/share/statusengine-ui

RUN apt-get update && apt-get install -y \
    git \
    libldap2-dev \
    libzip-dev \
    mysql-client \
    unzip \
    zip

RUN docker-php-ext-configure zip --with-libzip && \
    docker-php-ext-configure ldap --with-libdir=lib/x86_64-linux-gnu/ && \
    docker-php-ext-install ldap pdo pdo_mysql zip

RUN php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
RUN php -r "if (hash_file('SHA384', 'composer-setup.php') === '93b54496392c062774670ac18b134c3b3a95e5a5e5c8f1a9f115f203b75bf9a129d5daa8ba6a13e2cc8a1da0806388a8') { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('composer-setup.php'); } echo PHP_EOL;"
RUN php composer-setup.php
RUN php -r "unlink('composer-setup.php');"
RUN php composer.phar install

COPY ./etc/config.yml.docker /usr/share/statusengine-ui/etc/config.yml
