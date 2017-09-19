angular.module('Statusengine').directive('reloadConfig', function (ReloadService) {
    return {
        restrict: 'A',
        templateUrl: 'templates/directives/reloadconfig.html',
        scope: {},
        controller: function ($scope, $http) {
            $scope.do_auto_reload = ReloadService.getAutoReloadEnabled();
            $scope.ack_and_downtime_is_ok = ReloadService.getAckAndDowntimeIsOk();
            $scope.isLoggedIn = false;

            $scope.$watch('do_auto_reload', function () {
                ReloadService.setAutoReloadEnabled($scope.do_auto_reload);
            });

            $scope.$watch('ack_and_downtime_is_ok', function () {
                ReloadService.setAckAndDowntimeIsOk($scope.ack_and_downtime_is_ok);
            });

            $scope.checkLoginState = function(){
                $http.get("/api/index.php/loginstate", {}
                ).then(function (result) {
                    $scope.isLoggedIn = result.data.isLoggedIn;
                });
            };

            $scope.logout = function(){
                $http.get("/api/index.php/logout", {}
                ).then(function (result) {
                    $scope.isLoggedIn = false;
                    window.location = '/login.html';
                });
            };

            $scope.login = function(){
                $scope.isLoggedIn = false;
                var currentPage = location.pathname + window.location.hash;
                window.localStorage.setItem('lastPage', currentPage);
                window.location = '/login.html';
            };

            $scope.checkLoginState();

        }
    };
});