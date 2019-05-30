Statusengine Interface is able to read the configuration from a config file or environment variable.
If both is present the values from the configuration file gets used as preferred values.

| Priority | Source               | Comment                                                        |
|----------|----------------------|----------------------------------------------------------------|
| 0        | default value        | Hardcoded value                                                |
| 1        | environment variable | If present it will overwrite the default                       |
| 2        | configuration file   | If present it will overwrite the environment variable          |


To keep it simple it is recommended to define everything in the configuration file **or** the environment variables.
Even if its possible, I don't recommend to use a mix of both.

## List of available environment variables
| Environment variable                          | Type   | Required | Example / Comments                                                                      |
|-----------------------------------------------|--------|----------|-----------------------------------------------------------------------------------------|
| SEI_USE_CRATE                                 | bool   | yes      | You must set `SEI_USE_CRATE` or `SEI_USE_MYSQL`                                         |
| SEI_USE_MYSQL                                 | bool   | yes      |                                                                                         |
| SEI_MYSQL_HOST                                | string | depends  | Required if `SEI_USE_MYSQL` is enabled                                                  |
| SEI_MYSQL_PORT                                | int    | depends  | Required if `SEI_USE_MYSQL` is enabled                                                  |
| SEI_MYSQL_USER                                | string | depends  | Required if `SEI_USE_MYSQL` is enabled                                                  |
| SEI_MYSQL_PASSWORD                            | string | depends  | Required if `SEI_USE_MYSQL` is enabled                                                  |
| SEI_MYSQL_DATABASE                            | string | depends  | Required if `SEI_USE_MYSQL` is enabled                                                  |
| SEI_CRATE_NODES                               | array  | depends  | `export SEI_CRATE_NODES="127.0.0.1:4200,192.168.1.1:4200,192.168.10.1:4200"`            |
| SEI_ALLOW_ANONYMOUS                           | bool   | no       |                                                                                         |
| SEI_ANONYMOUS_CAN_SUBMIT_COMMANDS             | bool   | no       |                                                                                         |
| SEI_URLS_WITHOUT_LOGIN                        | array  | no       |                                                                                         |
| SEI_AUTH_TYPE                                 | string | no       |                                                                                         |
| SEI_LDAP_SERVER                               | string | depend   |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_PORT                                 | int    | no       |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_USE_SSL                              | bool   | no       |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_BIND_DN                              | string | no       |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_BIND_PASSWORD                        | string | depend   |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_BASE_DN                              | string | depend   |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_FILTER                               | string | no       |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_LDAP_ATTRIBUTE                            | string | no       |Required if `SEI_AUTH_TYPE` is `ldap`                                                    |
| SEI_DISPLAY_PERFDATA                          | bool   | no       |                                                                                         |
| SEI_PERFDATA_BACKEND                          | string | no       | On of `crate`, `graphite`, `mysql` or `elasticsearch`                                   |
| SEI_GRAPHITE_URL                              | string | depend   |Required if `SEI_DISPLAY_PERFDATA` is `1` and `SEI_PERFDATA_BACKEND` is `graphite`       |
| SEI_GRAPHITE_ILLEGAL_CHARACTERS               | string | no       |                                                                                         |
| SEI_GRAPHITE_PREFIX                           | string | no       |                                                                                         |
| SEI_GRAPHITE_USE_BASIC_AUTH                   | bool   | no       |                                                                                         |
| SEI_GRAPHITE_USER                             | string | depend   |Required if `SEI_GRAPHITE_USE_BASIC_AUTH` is `1`                                         |
| SEI_GRAPHITE_PASSWORD                         | string | depend   |Required if `SEI_GRAPHITE_USE_BASIC_AUTH` is `1`                                         |
| SEI_GRAPHITE_ALLOW_SELF_SIGNED_CERTIFICATES   | bool   | no       |                                                                                         |
| SEI_ELASTICSEARCH_INDEX                       | string | depend   |Required if `SEI_DISPLAY_PERFDATA` is `1` and `SEI_PERFDATA_BACKEND` is `elasticsearch`  |
| SEI_ELASTICSEARCH_ADDRESS                     | string | depend   |Required if `SEI_DISPLAY_PERFDATA` is `1` and `SEI_PERFDATA_BACKEND` is `elasticsearch`  |
| SEI_ELASTICSEARCH_PORT                        | int    | no       |                                                                                         |
| SEI_ELASTICSEARCH_PATTERN                     | string | depend   |                                                                                         |


## Default values
All variables have a predefined default value.
Search in the file [src/Config.php](/src/Config.php) for a variable name to get the default value.

## Documentation for each variable
More information about each variable can be found in
[etc/config.yml.example](/etc/config.yml.example).
Search for a variable without the `SEI_` prefix.

## Data types
| Data Type | How to pass                                      | Example                                                                      |
|-----------|--------------------------------------------------|------------------------------------------------------------------------------|
| string    | `VAR="value"`                                    | `export SEI_MYSQL_HOST="127.0.0.1"`                                          |
| int       | `VAR=value`                                      | `export SEI_MYSQL_PORT=3306`                                                 |
| bool      | `VAR=1` or out of `[1, true, on, 0, false, off]` | `export SEI_USE_MYSQL=1`                                                     |
| array     | `VAR=value1,value2`                              | `export SEI_CRATE_NODES="127.0.0.1:4200,192.168.1.1:4200,192.168.10.1:4200"` |


## Example

This examples work without any config.yml file.

#### Apache
````apacheconfig

<VirtualHost *:443>

    DocumentRoot "/usr/share/statusengine-ui/public/"

    SetEnv SEI_USE_MYSQL 1
    SetEnv SEI_MYSQL_HOST localhost
    SetEnv SEI_MYSQL_USER statusengine
    SetEnv SEI_MYSQL_PASSWORD password
    SetEnv SEI_MYSQL_DATABASE statusengine_data

    SetEnv SEI_DISPLAY_PERFDATA 1
    SetEnv SEI_PERFDATA_BACKEND mysql

    SetEnv SEI_ALLOW_ANONYMOUS 0

    SSLEngine On
    SSLCertificateFile    /etc/ssl/certs/ssl-cert-snakeoil.pem
    SSLCertificateKeyFile /etc/ssl/private/ssl-cert-snakeoil.key

    ErrorLog "/var/log/apache2/statusengine-ui-error.log"
    CustomLog "/var/log/apache2/statusengine-ui-access.log" combined
</VirtualHost>
````


#### Nginx
````
server {
    #Redirect http to https
    listen         80;

    server_tokens off;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;

    server_tokens off;
    ssl_certificate     /etc/ssl/certs/ssl-cert-snakeoil.pem;
    ssl_certificate_key /etc/ssl/private/ssl-cert-snakeoil.key;

    root   /usr/share/statusengine-ui/public/;
    index  index.html;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    location ~ \index.php {
        include /etc/nginx/fastcgi_params;
        fastcgi_pass     127.0.0.1:9000;
        fastcgi_index    index.php;
        fastcgi_param    SCRIPT_FILENAME $document_root/api/index.php;
        fastcgi_param    SCRIPT_NAME  /api/index.php;
        fastcgi_param    PHP_SELF     $document_uri;
        
        fastcgi_param    SEI_USE_MYSQL        1;
        fastcgi_param    SEI_MYSQL_HOST       localhost;
        fastcgi_param    SEI_MYSQL_USER       statusengine;
        fastcgi_param    SEI_MYSQL_PASSWORD   password;
        fastcgi_param    SEI_MYSQL_DATABASE   statusengine_data;
        
        fastcgi_param    SEI_DISPLAY_PERFDATA 1;
        fastcgi_param    SEI_PERFDATA_BACKEND mysql;
        
        fastcgi_param    SEI_ALLOW_ANONYMOUS  0;
    }

    location ~ /\.git {
        deny all;
    }

    # Remove css, js, and images from access log
    location ~* \.(?:css|js|svg|gif|png|html|ttf|ico|jpg|jpeg)$ {
        access_log off;
    }
}
````

### Bash
For Statusengine Interface Console (cli)

````
export SEI_USE_MYSQL=1
export SEI_MYSQL_HOST="localhost"
export SEI_MYSQL_USER="statusengine"
export SEI_MYSQL_PASSWORD="password"
export SEI_MYSQL_DATABASE="statusengine_data"

export SEI_DISPLAY_PERFDATA=1
export SEI_PERFDATA_BACKEND="mysql"

export SEI_ALLOW_ANONYMOUS=0

/usr/share/statusengine-ui/bin/Console.php users

````

### PHP built-in server

*Not supported* => https://bugs.php.net/bug.php?id=67808