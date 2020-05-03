angular.module('Login')
    .controller('IndexController', function ($scope, $http) {

        $scope.isAnonymousAllowed = false;

        $scope.csrf_name = null;
        $scope.csrf_value = null;

        $scope.username = '';
        $scope.password = '';

        $scope.wrongCredentials = false;

        $scope.lastPage = window.localStorage.getItem('lastPage');
        window.localStorage.removeItem('lastPage');

        $scope.fixBackground = function () {
            $('#background-image').backgrounder({element: 'body'});
            if ($(document).innerWidth() > 768) {
                var $loginBox = $('.login-box');
                $loginBox.css('top', Math.round($(document).innerHeight() / 2 - $loginBox.height() / 2) + 'px');
            }
        };

        $scope.checkLoginState = function () {
            $http.get("./api/index.php/loginstate", {}
            ).then(function (result) {
                $scope.isAnonymousAllowed = result.data.isAnonymousAllowed;
            });
        };

        $scope.getCsrf = function () {
            $http.get("./api/index.php/login").then(function (result) {
                    $scope.csrf_name = result.data.required_fields.csrf_name;
                    $scope.csrf_value = result.data.required_fields.csrf_value;

                    if (result.data.isLoggedIn === true) {
                        window.location = './';
                    }

                }
            );
        };

        $scope.submit = function () {
            $http.post("./api/index.php/login", {
                csrf_name: $scope.csrf_name,
                csrf_value: $scope.csrf_value,
                username: $scope.username,
                password: $scope.password
            }).then(function (result) {
                    if (result.data.hasOwnProperty('message')) {
                        if (result.data.message === 'Login Successfully') {
                            if ($scope.lastPage !== null) {
                                //window.location = $scope.lastPage;
                                window.location.href = $scope.lastPage;
                            } else {
                                //Browsers making strange thinks without the else
                                window.location = './';
                            }
                        }
                    }

                }, function errorCallback(response) {
                    $scope.wrongCredentials = true;
                    //Get new CSRF token for next post request
                    $scope.getCsrf();
                }
            );

        };

        $scope.fixBackground();
        $scope.getCsrf();
        $scope.checkLoginState();

    });