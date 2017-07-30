angular.module('Statusengine')
    .controller('IndexController', function ($scope, ReloadService) {
        $scope.reload = ReloadService.triggerReload;
    });