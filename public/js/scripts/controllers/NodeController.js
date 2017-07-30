angular.module('Statusengine')

    .controller("NodeController", function ($http, $interval, $scope, ReloadService, $document, $stateParams, $state) {
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;

        $scope.clusterNodes = [];
        $scope.cluster_filter = [];

        //default filter - show all host states
        $scope.state_filter = ['up', 'down', 'unreachable'];

        //chek url params
        if ($stateParams.show_state != '') {
            var tmpStateFilter = [];
            for (var key in $scope.state_filter) {
                tmpStateFilter.push(false);
                if ($scope.state_filter[key] == $stateParams.show_state) {
                    tmpStateFilter[key] = $stateParams.show_state;
                }
            }
            $scope.state_filter = tmpStateFilter;
        }

        $scope.hostname__like = '';

        var offset = 0;
        var limit = 50;

        $scope.reload = function () {
            offset = 0;
            $http.get("/api/index.php/hosts", {
                params: {
                    order: 'hostname',
                    direction: 'asc',
                    limit: limit,
                    offset: offset,
                    hostname__like: $scope.hostname__like,
                    "state[]": $scope.state_filter,
                    'cluster_name[]': $scope.cluster_filter
                }
            }).then(function (result) {
                    $scope.apiIsBusyOrNoDataAnymore = false;
                    $scope.data = result.data;
                }
            );

            $http.get("/api/index.php/cluster").then(function (result) {
                    $scope.clusterNodes = result.data;
                }
            );
        };

        $scope.loadMoreHosts = function () {
            $scope.apiIsBusyOrNoDataAnymore = true;
            offset += limit;

            $http.get("/api/index.php/hosts", {
                params: {
                    order: 'hostname',
                    direction: 'asc',
                    limit: limit,
                    offset: offset,
                    hostname__like: $scope.hostname__like,
                    "state[]": $scope.state_filter,
                    'cluster_name[]': $scope.cluster_filter
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
            if(!$state.is('nodes')){
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
        $scope.$watch('[hostname__like, state_filter, cluster_filter]', function(){
            offset = 0;
            $scope.data =[];
            $scope.reload();
        }, true);

        $scope.loadMoreHostsOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreHosts();
            }
        };

    });