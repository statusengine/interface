angular.module('Statusengine').directive('hoststatus', function(){
    return {
        restrict: 'E',
        templateUrl: 'templates/directives/hoststatus.html',
        scope: {'hoststatusOverview':'='},
        controller: function($scope){

        }
    };
});