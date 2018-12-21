angular.module('Statusengine').directive('reloadConfig', function(ReloadService, NightModeService){
    return {
        restrict: 'A',
        templateUrl: 'templates/directives/reloadconfig.html',
        scope: {},
        controller: function($scope, $http){

            $('#pageSettings').click(function(event){
                event.stopPropagation();
            });

            $scope.do_auto_reload = ReloadService.getAutoReloadEnabled();
            $scope.ack_and_downtime_is_ok = ReloadService.getAckAndDowntimeIsOk();
            $scope.autoreload_frequency = String(ReloadService.getAutoReloadFrequency()); //String() cast fixe selected
            $scope.isLoggedIn = false;
            $scope.nightMode = NightModeService.isNightModeEnabled();
            if($scope.nightMode){
                $('body').addClass('night-mode');
            }


            $scope.$watch('nightMode', function(){
                if($scope.nightMode === false){
                    $('body').removeClass('night-mode');
                    NightModeService.disableNightMode();
                }

                if($scope.nightMode === true){
                    $('body').addClass('night-mode');
                    NightModeService.enableNightMode();
                }
            });

            $scope.$watch('do_auto_reload', function(){
                ReloadService.setAutoReloadEnabled($scope.do_auto_reload);
            });

            $scope.$watch('ack_and_downtime_is_ok', function(){
                ReloadService.setAckAndDowntimeIsOk($scope.ack_and_downtime_is_ok);
            });

            $scope.$watch('autoreload_frequency', function(){
                ReloadService.setAutoReloadFrequency($scope.autoreload_frequency);
            });


            $scope.checkLoginState = function(){
                $http.get("/api/index.php/loginstate", {}
                ).then(function(result){
                    $scope.isLoggedIn = result.data.isLoggedIn;
                });
            };

            $scope.logout = function(){
                $http.get("/api/index.php/logout", {}
                ).then(function(result){
                    $scope.isLoggedIn = false;
                    window.location = '/login.html';
                });
            };

            $scope.login = function(){
                $scope.isLoggedIn = false;
                var currentPage = location.pathname + window.location.hash;
                window.localStorage.setItem('lastPage', currentPage);
                window.location = '/login.html';
            };

            $scope.checkLoginState();

        }
    };
});