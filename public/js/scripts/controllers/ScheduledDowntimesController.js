angular.module('Statusengine')

    .controller("ScheduleddowntimesController", function($http, $interval, $scope, ReloadService, $document, $state){

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
        if(window.localStorage.getItem('scheduleddowntimes_object_type') == 'service'){
            $scope.object_type = 'service';
        }


        var offset = 0;
        var limit = 50;

        $scope.reload = function(){
            offset = 0;

            if($scope.object_type == 'host'){
                window.localStorage.removeItem('scheduleddowntimes_object_type');
            }else{
                window.localStorage.setItem('scheduleddowntimes_object_type', 'service');
            }

            $http.get("./api/index.php/scheduleddowntimes", {
                params: {
                    object_type: $scope.object_type,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
                    'cluster_name[]': $scope.cluster_filter,
                    limit: limit,
                    offset: offset
                }
            }).then(function(result){
                    $scope.moreDataAvailable = true;
                    $scope.apiIsBusyOrNoDataAnymore = false;
                    $scope.data = result.data;
                    if(result.data.length === 0){
                        $scope.moreDataAvailable = false;
                    }
                }
            );

            $http.get("./api/index.php/cluster").then(function(result){
                    $scope.clusterNodes = result.data;
                }
            );
        };

        $scope.loadMoreDowntimes = function(){
            $scope.apiIsBusyOrNoDataAnymore = true;
            offset += limit;

            $http.get("./api/index.php/scheduleddowntimes", {
                params: {
                    object_type: $scope.object_type,
                    hostname__like: $scope.hostname__like,
                    servicedescription__like: $scope.servicedescription__like,
                    'cluster_name[]': $scope.cluster_filter,
                    limit: limit,
                    offset: offset
                }
            }).then(function(result){
                    angular.forEach(result.data, function(item){
                        $scope.data.push(item);
                    });

                    $scope.apiIsBusyOrNoDataAnymore = false;
                    if(result.data.length === 0){
                        $scope.moreDataAvailable = false;
                        $scope.apiIsBusyOrNoDataAnymore = true;
                    }
                }
            );
        };

        $document.on('scroll', function(){
            if(!$state.is('scheduleddowntimes')){
                return;
            }
            //Disable or enable auto reload - so you can read the logs
            if(hasUserAutoReloadEnabled === true){
                if($document.scrollTop() < 10){
                    ReloadService.setAutoReloadEnabledTemporary(true);
                }else{
                    ReloadService.setAutoReloadEnabledTemporary(false);
                }
            }
        });

        $scope.submitDeleteHostDowntime = function(downtime, cancelServiceDowntimes){
            if($scope.isAllowedToSubmitCommand === false){
                return;
            }

            if(cancelServiceDowntimes === false){
                $scope.submitDeleteDowntime(downtime.internal_downtime_id, downtime.node_name);
                return;
            }

            var data = {
                downtime_id: downtime.internal_downtime_id,
                node_name: downtime.node_name
            };

            $http.get("./api/index.php/delete_host_and_service_downtimes", {
                params: data
            }).then(function(result){
                noty({
                    theme: 'metrouiAdminLTE',
                    progressBar: true,
                    layout: 'bottomRight',
                    type: 'success',
                    text: 'Command was sent to Statusengine task queue',
                    timeout: 2500,
                    animation: {
                        open: 'animated flipInX',
                        close: 'animated flipOutX'
                    }

                });
            });
        };


        $scope.submitDeleteDowntime = function(internal_downtime_id, node_name){
            if($scope.isAllowedToSubmitCommand === false){
                return;
            }

            var command = 'DEL_HOST_DOWNTIME';
            if($scope.object_type == 'service'){
                command = 'DEL_SVC_DOWNTIME';
            }

            var data = {};
            data['command_name'] = command;
            data['downtime_id'] = internal_downtime_id;
            data['node_name'] = node_name;

            $http.get("./api/index.php/externalcommand_args", {
                params: data
            }).then(function(result){
                noty({
                    theme: 'metrouiAdminLTE',
                    progressBar: true,
                    layout: 'bottomRight',
                    type: 'success',
                    text: 'Command was sent to Statusengine task queue',
                    timeout: 2500,
                    animation: {
                        open: 'animated flipInX',
                        close: 'animated flipOutX'
                    }

                });
            });
        };

        $scope.getLoginState = function(){
            $http.get("./api/index.php/loginstate", {
                params: {}
            }).then(function(result){
                $scope.isAllowedToSubmitCommand = false;
                if(result.data.canAnonymousSubmitCommand === true || result.data.isLoggedIn === true){
                    $scope.isAllowedToSubmitCommand = true;
                }
            });
        };


        ReloadService.setCallback($scope.reload);

        //triggers $scope.reload() on load and on search
        $scope.$watch('[hostname__like, servicedescription__like, cluster_filter, object_type]', function(){
            $scope.reload();
        }, true);


        $scope.loadMoreScheduledDowntimesOnScroll = function(){
            if($scope.apiIsBusyOrNoDataAnymore === false){
                $scope.loadMoreDowntimes();
            }
        };

        $scope.getLoginState();

    });