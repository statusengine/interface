angular.module('Statusengine')

    .controller("ProblemController", function ($http, $interval, $scope, ReloadService, $document, $stateParams, $state) {
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.clusterNodes = [];
        $scope.cluster_filter = [];

        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;

        $scope.tiles = false;
        if (window.localStorage.getItem('problemsAsTiles') == 'true') {
            $scope.tiles = true;
        }


        $scope.hostname__like = '';
        $scope.servicedescription__like = '';

        var offset = 0;
        var limit = 50;

        $scope.reload = function () {
            offset = 0;
            $http.get("/api/index.php/problems", {
                params: {
                    order: 'hostname,service_description',
                    direction: 'asc',
                    limit: limit,
                    offset: offset,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
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

            $http.get("/api/index.php/problems", {
                params: {
                    order: 'hostname,service_description',
                    direction: 'asc',
                    limit: limit,
                    offset: offset,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
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

        $scope.changeView = function () {


            if ($scope.tiles === true) {
                $scope.tiles = false;
                window.localStorage.removeItem('problemsAsTiles');
            } else {
                $scope.tiles = true;
                window.localStorage.setItem('problemsAsTiles', true);
            }
        }

        $document.on('scroll', function () {
            if (!$state.is('problems')) {
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
        $scope.$watch('[servicedescription__like, hostname__like, cluster_filter]', function () {
            offset = 0;
            $scope.reload();
        }, true);


        $scope.loadMoreHostsOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreHosts();
            }
        };
    });
