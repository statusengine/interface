angular.module('Statusengine')

    .controller("ServiceDetailsController", function($http, $interval, $scope, ReloadService, $stateParams, $filter, $uibModal){
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.serviceNotFound = false;

        $scope.nodename = decodeURI($stateParams.nodename);
        $scope.servicedescription = decodeURI($stateParams.servicedescription);

        $scope.isAllowedToSubmitCommand = false;

        $scope.apiIsBusyOrNoDataAnymore = true;
        $scope.timespan = 9000;
        $scope.algorithm = 'avg';

        $scope.external_urls = {};

        $scope.reload = function(){
            offset = 0;
            $http.get("./api/index.php/servicedetails", {
                params: {
                    hostname: $scope.nodename,
                    servicedescription: $scope.servicedescription
                }
            }).then(function(result){
                    $scope.apiIsBusyOrNoDataAnymore = false;
                    $scope.data = result.data;

                    if(result.data.hasOwnProperty('external_urls')){
                        $scope.external_urls = result.data.external_urls;
                    }

                    if(!result.data.servicestatus.hasOwnProperty('service_description')){
                        $scope.serviceNotFound = true;
                        $('#service-not-found-modal').modal('show');
                    }

                    if(result.data.servicestatus.hasOwnProperty('service_description') > 0 && $scope.serviceNotFound === true){
                        $scope.serviceNotFound = false;
                        $('#service-not-found-modal-dialog').addClass('animated hinge');
                        $('#service-not-found-modal').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function(){
                            $('#service-not-found-modal').modal('hide');
                            $('#service-not-found-modal-dialog').removeClass('animated hinge');
                        });
                    }

                    if(result.data.servicestatus.hasOwnProperty('problem_has_been_acknowledged')){
                        if(result.data.servicestatus.problem_has_been_acknowledged){
                            $scope.loadAcknowledgementData();
                        }else{
                            $scope.acknowledgementData = null;
                        }
                    }

                    if(result.data.servicestatus.hasOwnProperty('scheduled_downtime_depth')){
                        if(result.data.servicestatus.scheduled_downtime_depth > 0){
                            $scope.loadServiceDowntimeData();
                        }else{
                            $scope.downtimeData = null;
                        }
                    }

                }
            );
        };

        $scope.loadAcknowledgementData = function(){
            $http.get("./api/index.php/serviceacknowledgements", {
                params: {
                    hostname: $scope.nodename,
                    servicedescription: $scope.servicedescription,
                    limit: 1,
                    order: 'entry_time',
                    direction: 'desc'
                }
            }).then(function(result){
                if(result.data.length > 0){
                    $scope.acknowledgementData = result.data[0];
                }

            });
        };

        $scope.loadServiceDowntimeData = function(){
            $http.get("./api/index.php/servicedowntime", {
                params: {
                    hostname: $scope.nodename,
                    servicedescription: $scope.servicedescription
                }
            }).then(function(result){
                if(result.data.length > 0){
                    $scope.downtimeData = result.data[0];
                }

            });
        };

        $scope.submitDeleteDowntime = function(internal_downtime_id){
            if($scope.isAllowedToSubmitCommand === false){
                return;
            }
            var data = {};
            data['command_name'] = 'DEL_SVC_DOWNTIME';
            data['downtime_id'] = internal_downtime_id;
            data['node_name'] = $scope.data.servicestatus.node_name;

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

        $scope.submitPassiveCheckResult = function(){
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitPassiveServiceCheckResult.html',
                controller: 'SubmitPassiveServiceCheckController'
            });

            modal.result.then(function(data){
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.sendCustomServiceNotification = function(){
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitCustomServiceNotification.html',
                controller: 'SubmitCustomServiceNotificationController'
            });

            modal.result.then(function(data){
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.submitServiceDowntime = function(){
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitServiceDowntime.html',
                controller: 'SubmitServiceDowntimeController'
            });

            modal.result.then(function(data){
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.submitServiceAcknowledgement = function(){
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitServiceAcknowledgement.html',
                controller: 'SubmitServiceAcknowledgementController'
            });

            modal.result.then(function(data){
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.sendCommandWithArgs = function(data){
            if($scope.isAllowedToSubmitCommand === false){
                return;
            }
            data['hostname'] = $scope.data.servicestatus.hostname;
            data['servicedescription'] = $scope.data.servicestatus.service_description;
            data['node_name'] = $scope.data.servicestatus.node_name;

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

        $scope.sendCommand = function(commandString){
            if($scope.isAllowedToSubmitCommand === false){
                return;
            }
            /**
             * @see src/Generators/ExternalCommand.php for the commandId values
             */

            var commandId = null;
            switch(commandString){
                case 'notifications':
                    commandId = 11; //Enable
                    if($scope.data.servicestatus.notifications_enabled === true){
                        commandId = 10; //Disable
                    }
                    break;

                case 'activeChecks':
                    commandId = 17; //Enable
                    if($scope.data.servicestatus.active_checks_enabled === true){
                        commandId = 16; //Disable
                    }
                    break;

                case 'passiveChecks':
                    commandId = 19; //Enable
                    if($scope.data.servicestatus.passive_checks_enabled === true){
                        commandId = 18; //Disable
                    }
                    break;

                case 'flappDetection':
                    commandId = 13; //Enable
                    if($scope.data.servicestatus.flap_detection_enabled === true){
                        commandId = 12; //Disable
                    }
                    break;

                case 'eventHandler':
                    commandId = 15; //Enable
                    if($scope.data.servicestatus.event_handler_enabled === true){
                        commandId = 14; //Disable
                    }
                    break;

                case 'reschedule':
                    commandId = 1;
                    break;

                default:
                    console.error('Unknown command');
                    return;
                    break;
            }

            $http.get("./api/index.php/externalcommand", {
                params: {
                    hostname: $scope.data.servicestatus.hostname,
                    servicedescription: $scope.data.servicestatus.service_description,
                    node_name: $scope.data.servicestatus.node_name,
                    command: commandId
                }
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

        $scope.getLoginState();
        $scope.reload();
    });
