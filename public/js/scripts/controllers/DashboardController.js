angular.module('Statusengine')

    .controller("DashboardController", function ($http, $interval, $scope, ReloadService) {
        $scope.reload = function () {
            $http.get("/api/index.php").then(function (result) {
                    $scope.data = result.data;
                }
            );
        };

        $scope.reload();

        ReloadService.enableAutoloadIfRequired();
        ReloadService.setCallback($scope.reload);

    });