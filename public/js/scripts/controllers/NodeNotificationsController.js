angular.module('Statusengine')

    .controller("NodeNotificationsController", function ($http, $interval, $scope, ReloadService, $document, $stateParams, $state) {
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;

        $scope.nodename = decodeURI($stateParams.nodename);

        //default filter - show all host states
        $scope.state_filter = ['up', 'down', 'unreachable'];

        $scope.output__like = '';

        var offset = 0;
        var limit = 50;

        $scope.reload = function () {
            offset = 0;
            $http.get("/api/index.php/hostnotifications", {
                params: {
                    hostname: $scope.nodename,
                    order: 'start_time',
                    direction: 'desc',
                    limit: limit,
                    offset: offset,
                    output__like: $scope.output__like,
                    "state[]": $scope.state_filter
                }
            }).then(function (result) {
                    $scope.apiIsBusyOrNoDataAnymore = false;
                    $scope.data = result.data;
                }
            );

        };

        $scope.loadMoreHosts = function () {
            $scope.apiIsBusyOrNoDataAnymore = true;
            offset += limit;

            $http.get("/api/index.php/hostnotifications", {
                params: {
                    hostname: $scope.nodename,
                    order: 'start_time',
                    direction: 'desc',
                    limit: limit,
                    offset: offset,
                    output__like: $scope.output__like,
                    "state[]": $scope.state_filter
                }
            }).then(function (result) {
                    angular.forEach(result.data, function (item) {
                        $scope.data.push(item);
                    });

                    $scope.apiIsBusyOrNoDataAnymore = false;
                    if (result.data.length === 0) {
                        $scope.moreDataAvailable = false;
                        $scope.apiIsBusyOrNoDataAnymore = true;
                    }
                }
            );
        };

        $document.on('scroll', function () {
            if(!$state.is('nodenotifications')){
                return;
            }
            //Disable or enable auto reload - so you can read the logs
            if (hasUserAutoReloadEnabled === true) {
                if ($document.scrollTop() < 10) {
                    ReloadService.setAutoReloadEnabledTemporary(true);
                } else {
                    ReloadService.setAutoReloadEnabledTemporary(false);
                }
            }
        });

        ReloadService.setCallback($scope.reload);

        //triggers reload on load and on search
        $scope.$watch('[output__like, state_filter]', function(){
            offset = 0;
            $scope.reload();
        }, true);

        $scope.loadMoreHostsOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreHosts();
            }
        };

    });