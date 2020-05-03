angular.module('Statusengine')

    .controller("LogentryController", function ($http, $interval, $scope, ReloadService, $document, $state) {

        ReloadService.enableAutoloadIfRequired();
        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();
        $scope.clusterNodes = [];
        $scope.cluster_filter = [];

        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;
        $scope.logentry_data__like = '';

        $scope.reload = function () {
            $http.get("./api/index.php/logentries", {
                params: {
                    limit: 50,
                    logentry_data__like: $scope.logentry_data__like,
                    'cluster_name[]': $scope.cluster_filter
                }
            }).then(function (result) {
                    $scope.moreDataAvailable = true;
                    $scope.apiIsBusyOrNoDataAnymore = false;
                    $scope.data = result.data;
                    if (result.data.length === 0) {
                        $scope.moreDataAvailable = false;
                    }
                }
            );

            $http.get("./api/index.php/cluster").then(function (result) {
                    $scope.clusterNodes = result.data;
                }
            );
        };

        $scope.loadMoreLogentries = function () {
            $scope.apiIsBusyOrNoDataAnymore = true;

            $http.get("./api/index.php/logentries", {
                params: {
                    limit: 50,
                    entry_time__lt: $scope.data[$scope.data.length - 1].entry_time,
                    logentry_data__like: $scope.logentry_data__like,
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
            if(!$state.is('logentries')){
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

        //triggers $scope.reload() on load and on search
        $scope.$watch('[logentry_data__like, cluster_filter]', $scope.reload, true);

        $scope.loadMoreLogentriesOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreLogentries();
            }
        };

    });