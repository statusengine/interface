angular.module('Statusengine').directive('hoststatuspill', function () {
    return {
        restrict: 'E',
        templateUrl: 'templates/directives/hoststatuspill.html',
        scope: {'hoststatus': '='},
        controller: function ($scope) {

            $scope.stateClass = 'bg-primary';

            $scope.colorByState = function () {
                $scope.stateClass = 'bg-primary';

                if ($scope.state == 0) {
                    $scope.stateClass = 'bg-green';
                }

                if ($scope.state == 1) {
                    $scope.stateClass = 'bg-red';
                }
            };

            $scope.$watch('hoststatus', function () {
                if($scope.hoststatus.hasOwnProperty('current_state')){
                    $scope.state = $scope.hoststatus.current_state;
                }

                if($scope.hoststatus.hasOwnProperty('state')){
                    $scope.state = $scope.hoststatus.state;
                }

                $scope.colorByState();
            });

        }
    };
});