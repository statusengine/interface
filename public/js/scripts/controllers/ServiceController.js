angular.module('Statusengine')

    .controller("ServiceController", function ($http, $interval, $scope, ReloadService, $document, $stateParams, $state) {
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.clusterNodes = [];
        $scope.cluster_filter = [];


        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;

        //default filter - show all host states
        $scope.state_filter = ['ok', 'warning', 'critical', 'unknown'];

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
        $scope.servicedescription__like = '';

        var offset = 0;
        var limit = 50;

        $scope.reload = function () {
            offset = 0;
            $http.get("/api/index.php/services", {
                params: {
                    order: 'hostname,service_description',
                    direction: 'asc',
                    limit: limit,
                    offset: offset,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
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

            $http.get("/api/index.php/services", {
                params: {
                    order: 'hostname,service_description',
                    direction: 'asc',
                    limit: limit,
                    offset: offset,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
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
            if (!$state.is('services')) {
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
        $scope.$watch('[servicedescription__like, hostname__like, state_filter, cluster_filter]', function () {
            offset = 0;
            $scope.reload();
        }, true);


        $scope.loadMoreHostsOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreHosts();
            }
        };

    });
