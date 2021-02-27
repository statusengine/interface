angular.module('Statusengine').directive('nodesearch', function ($http, $timeout) {
    return {
        restrict: 'A',
        templateUrl: 'templates/directives/nodesearch.html',
        scope: {},
        controller: function ($scope) {

            $scope.nodename = '';
            $scope.searching = false;
            $scope.result = [];
            $scope.noResult = false;
            $scope.emptySearchString = true;

            var initializing = true;


            $scope.search = function () {
                $scope.searching = true;

                $http.get("./api/index.php/hostsearch", {
                    params: {
                        limit: 25,
                        hostname__like: $scope.nodename
                    }
                }).then(function (result) {
                    $scope.result = result.data;

                    $scope.noResult = false;
                    if($scope.result.length == 0){
                        $scope.noResult = true;
                    }

                    $timeout(function () {
                        //Let the weel spin for one seconds
                        $scope.searching = false;
                    }, 500);
                });

            };

            $scope.$watch('nodename', function () {
                $scope.result = [];
                if($scope.nodename == ''){
                    $scope.emptySearchString = true;
                    $scope.noResult = false;
                }else{
                    $scope.emptySearchString = false;
                }

                if (initializing === false && $scope.emptySearchString == false) {
                    $scope.search();
                }

                initializing = false;
            });


        }
    };
});
