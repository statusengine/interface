angular.module('Statusengine')
    .controller('IndexController', function ($scope, ReloadService, $http, $interval) {
        $scope.reload = ReloadService.triggerReload;

        $scope.date = {};
        $scope.acknowledgements = 0;
        $scope.downtimes = 0;

        $scope.reload = function () {
            $http.get("/api/index.php/menustats").then(function (result) {
                    $scope.data = result.data;
                    $scope.acknowledgements = $scope.data.number_of_host_acknowledgements;
                    $scope.acknowledgements += $scope.data.number_of_service_acknowledgements;

                    $scope.downtimes = $scope.data.number_of_scheduled_host_downtimes;
                    $scope.downtimes += $scope.data.number_of_scheduled_service_downtimes;
                }
            );
        };


        $interval($scope.reload, 30000);
        $scope.reload();
    });