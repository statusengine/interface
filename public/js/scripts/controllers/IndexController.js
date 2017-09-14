angular.module('Statusengine')
    .controller('IndexController', function ($scope, ReloadService, $http, $interval, $timeout) {
        $scope.reload = ReloadService.triggerReload;

        $scope.date = {};
        $scope.acknowledgements = 0;
        $scope.downtimes = 0;

        $scope.loadMenustats = function () {
            $('#menustatsRefresh').show();
            $http.get("/api/index.php/menustats").then(function (result) {
                    $scope.data = result.data;
                    $scope.acknowledgements = $scope.data.number_of_host_acknowledgements;
                    $scope.acknowledgements += $scope.data.number_of_service_acknowledgements;

                    $scope.downtimes = $scope.data.number_of_scheduled_host_downtimes;
                    $scope.downtimes += $scope.data.number_of_scheduled_service_downtimes;
                    $timeout(function () {
                        //Let the weel spin for one seconds
                        $('#menustatsRefresh').hide();
                    }, 1000);
                }
            );
        };


        $interval($scope.loadMenustats, 30000);
        $scope.loadMenustats();
    });