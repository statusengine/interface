angular.module('Statusengine').directive('deleteHostDowntime', function () {
    return {
        restrict: 'E',
        templateUrl: 'templates/directives/delete_host_downtime.html',

        // Notice!
        // This directive has no own $scope and shares the $scope with the controller/template where the directive is used.
        controller: function ($scope) {

            var downtimeData = {};
            var callbackName = false;

            $scope.deleteServiceDowntimes = false;

            $scope.setHostDowntimeDeleteCallback = function(_callback){
                callbackName = _callback;
            };

            $scope.setHostForDowntimeDelete = function(_downtimeData){
                downtimeData = _downtimeData;
            };

            $scope.triggerHostDowntimeDeleteCallback = function(){
                if(callbackName === false){
                    console.warn('No callback given!')
                }else{
                    $scope[callbackName](downtimeData, $scope.deleteServiceDowntimes);
                }
                $('#delete-host-downtime-modal').modal('hide');
            }

        },

        link: function($scope, element, attr){

            $scope.confirmDeleteHostDowntime = function(downtimeData){

                if(attr.hasOwnProperty('callback')){
                    $scope.setHostDowntimeDeleteCallback(attr.callback);
                }

                $scope.setHostForDowntimeDelete(downtimeData);

                $('#delete-host-downtime-modal').modal('show');

            };

        }
    };
});