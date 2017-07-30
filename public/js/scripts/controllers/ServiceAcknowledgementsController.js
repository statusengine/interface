angular.module('Statusengine')

    .controller("ServiceAcknowledgementsController", function ($http, $interval, $scope, ReloadService, $document, $stateParams, $state) {
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.moreDataAvailable = true;
        $scope.apiIsBusyOrNoDataAnymore = true;

        $scope.nodename = decodeURI($stateParams.nodename);
        $scope.servicedescription = decodeURI($stateParams.servicedescription);


        //default filter - show all host states
        $scope.state_filter = ['ok', 'warning', 'critical', 'unknown'];

        $scope.comment_data__like = '';

        var offset = 0;
        var limit = 50;

        $scope.reload = function () {
            offset = 0;
            $http.get("/api/index.php/serviceacknowledgements", {
                params: {
                    hostname: $scope.nodename,
                    servicedescription: $scope.servicedescription,
                    order: 'entry_time',
                    direction: 'desc',
                    limit: limit,
                    offset: offset,
                    comment_data__like: $scope.comment_data__like,
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

            $http.get("/api/index.php/serviceacknowledgements", {
                params: {
                    hostname: $scope.nodename,
                    servicedescription: $scope.servicedescription,
                    order: 'entry_time',
                    direction: 'desc',
                    limit: limit,
                    offset: offset,
                    comment_data__like: $scope.comment_data__like,
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
            if(!$state.is('serviceacknowledgements')){
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
        $scope.$watch('[comment_data__like, state_filter]', function(){
            offset = 0;
            $scope.reload();
        }, true);

        $scope.loadMoreHostsOnScroll = function () {
            if ($scope.apiIsBusyOrNoDataAnymore === false) {
                $scope.loadMoreHosts();
            }
        };

    });