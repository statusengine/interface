angular.module('Statusengine').directive('globalproblem', function ($http, $interval) {
    return {
        restrict: 'A',
        templateUrl: 'templates/directives/globalproblem.html',
        scope: {},
        controller: function ($scope) {

            $scope.issueCount = 0;
            $scope.services = [];

            var refreshTimer = null;


            $scope.load = function () {
                $http.get("./api/index.php/globalproblems", {
                    params: {
                        limit: 5
                    }
                }).then(function (result) {
                    $scope.issueCount = result.data.problemCount;

                    $scope.services = result.data.services;
                });

            };

            refreshTimer = $interval($scope.load, 60000);
            $scope.load();
        }
    };
});
