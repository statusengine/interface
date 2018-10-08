# Statusengine UI
Statusengine UI is a lightweight, responsive, modern web interface, you can use, to make your monitoring data visable.

It's build with AngularJS, communicating via a JSON API to be easy extendable and also has
simple support to render performance data.

Visit the [documentation](https://statusengine.org/) for more information about Statusengine UI

## Install
````
apt-get install php-mysql php-ldap
php5enmod ldap
service apache2 restart
composer install
````

## Web server
Point the document root of your web server to the `public` folder

## Create user to login
To login, first of all you need to create a new user.

This example will show you, of how to create a new user:

`bin/Console.php users add --username=admin --password=admin`

Username: `admin`

Password: `admin`

Run `bin/Console.php users --help` to get more information.

You can also run the command without `--username` and `--password` to keep your password private and hidden from history.

Run `bin/Console.php users` to get a list of all users:

````
+----------+--------------------------------------------------------------+
| Username | Password                                                     |
+----------+--------------------------------------------------------------+
| admin    | $2y$10$PzJ8sqeww/.eBz1xfiBCzeuBa5Y9ZwufEtElPt0QqlmlYNEXfDzK6 |
| foobar   | $2y$10$DQnPleIFrffDv3b6q3TeBei3oMju9n/C/m1KF.//IUnT2lDCOy/QG |
+----------+--------------------------------------------------------------+
````

## Docker Usage

The configuration file used by docker-compose is located in `etc/config.yml.docker`.

Spin up the containers and detach:
```
docker-compose up -d
```

Initialize the database with sample data:
```
docker exec statusengine_db bash /tmp/init_db.sh
```

Create a new user:
```
docker exec statusengine_ui php /usr/share/statusengine-ui/bin/Console.php users add --username \"admin\" --password \"admin\"
```

## License
GNU General Public License v3.0
````
Statusengine UI
Copyright (C) 2016-2018  Daniel Ziegler

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
`````

## Used `bower` components
To avoid compatibility issues, all `bower` components are pushed to the repository.

| Name                                                                                  | License                                                                         |
|---------------------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| [AngularJS](https://github.com/angular/angular.js)                                    | MIT License                                                                     |
| [angular-bootstrap](http://angular-ui.github.io/bootstrap/)                           | MIT License                                                                     |
| [angular-scroll](https://github.com/oblador/angular-scroll)                           | MIT License                                                                     |
| [angular-ui-router](https://github.com/angular-ui/angular-ui-router-bower)            | MIT License                                                                     |
| [animate.css](https://github.com/daneden/animate.css)                                 | MIT License                                                                     |
| [bootstrap](https://github.com/twbs/bootstrap)                                        | MIT License                                                                     |
| [chart.js](https://github.com/chartjs/Chart.js)                                       | MIT License                                                                     |
| [font-awesome](https://github.com/FortAwesome/Font-Awesome)                           | [Font-Awesome#license](https://github.com/FortAwesome/Font-Awesome#license)     |
| [jquery](https://github.com/jquery/jquery)                                            | [https://jquery.org/license/](https://jquery.org/license/)                      |
| [ngInfiniteScroll](https://github.com/ng-infinite-scroll/ng-infinite-scroll-bower)    | MIT License                                                                     |
| [noty](https://github.com/needim/noty)                                                | MIT License                                                                     |
| [jQuery-Backgrounder](https://github.com/bigfolio/jQuery-Backgrounder)                | MIT License                                                                     |

