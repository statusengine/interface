angular.module('Statusengine')

    .controller("AcknowledgementsController", function ($http, $interval, $scope, ReloadService, $document, $state) {

        ReloadService.enableAutoloadIfRequired();
        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();
        $scope.clusterNodes = [];
        $scope.cluster_filter = [];

        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;
        $scope.hostname__like = '';
        $scope.servicedescription__like = '';

        $scope.isAllowedToSubmitCommand = false;
        $scope.object_type = 'host';
        if (window.localStorage.getItem('acknowledgements_object_type') == 'service') {
            $scope.object_type = 'service';
        }


        var offset = 0;
        var limit = 50;

        $scope.reload = function () {
            offset = 0;

            if ($scope.object_type == 'host') {
                window.localStorage.removeItem('acknowledgements_object_type');
            } else {
                window.localStorage.setItem('acknowledgements_object_type', 'service');
            }

            $http.get("/api/index.php/acknowledgements", {
                params: {
                    object_type: $scope.object_type,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
                    'cluster_name[]': $scope.cluster_filter,
                    limit: limit,
                    offset: offset
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
        };

        $scope.loadMoreAcknowledgements = function () {
            $scope.apiIsBusyOrNoDataAnymore = true;
            offset += limit;

            $http.get("/api/index.php/acknowledgements", {
                params: {
                    object_type: $scope.object_type,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
                    'cluster_name[]': $scope.cluster_filter,
                    limit: limit,
                    offset: offset
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
            if (!$state.is('acknowledgements')) {
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

        $http.get("/api/index.php/cluster").then(function (result) {
                $scope.clusterNodes = result.data;
            }
        );


        ReloadService.setCallback($scope.reload);

        //triggers $scope.reload() on load and on search
        $scope.$watch('[hostname__like, servicedescription__like, object_type, cluster_filter]', function () {
            $scope.reload();
        }, true);


        $scope.loadMoreScheduledDowntimesOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreAcknowledgements();
            }
        };


    });