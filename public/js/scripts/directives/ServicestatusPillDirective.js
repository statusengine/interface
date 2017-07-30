angular.module('Statusengine').directive('servicestatuspill', function () {
    return {
        restrict: 'E',
        templateUrl: 'templates/directives/servicestatuspill.html',
        scope: {'servicestatus': '='},
        controller: function ($scope) {

            $scope.stateClass = 'bg-primary';

            $scope.colorByState = function () {
                $scope.stateClass = 'bg-primary';

                if ($scope.state == 0) {
                    $scope.stateClass = 'bg-green';
                }

                if ($scope.state == 2) {
                    $scope.stateClass = 'bg-red';
                }

                if ($scope.state == 1) {
                    $scope.stateClass = 'bg-yellow';
                }
            };

            $scope.$watch('servicestatus', function () {
                if ($scope.servicestatus.hasOwnProperty('current_state')) {
                    $scope.state = $scope.servicestatus.current_state;
                }

                if ($scope.servicestatus.hasOwnProperty('state')) {
                    $scope.state = $scope.servicestatus.state;
                }

                $scope.colorByState();
            });

        }

    }

});