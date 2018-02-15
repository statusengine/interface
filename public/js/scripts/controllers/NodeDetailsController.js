angular.module('Statusengine')

    .controller("NodeDetailsController", function ($http, $interval, $scope, ReloadService, $stateParams, $location, $uibModal) {
        ReloadService.enableAutoloadIfRequired();

        var hasUserAutoReloadEnabled = ReloadService.getAutoReloadEnabled();

        $scope.hostNotFound = false;

        $scope.nodename = decodeURI($stateParams.nodename);

        $scope.isAllowedToSubmitCommand = false;

        $scope.apiIsBusyOrNoDataAnymore = true;

        //default filter - show all service states
        $scope.state_filter = ['ok', 'warning', 'critical', 'unknown'];
        $scope.servicedescription__like = '';

        $scope.init = true;

        $scope.reload = function () {
            offset = 0;
            $http.get("/api/index.php/hostdetails", {
                params: {
                    hostname: $scope.nodename,
                    servicedescription__like: $scope.servicedescription__like,
                    "service_state[]": $scope.state_filter
                }
            }).then(function (result) {
                    $scope.data = result.data;
                    $scope.init = false;

                    if (!result.data.hoststatus.hasOwnProperty('hostname')) {
                        $scope.hostNotFound = true;
                        $('#host-not-found-modal').modal('show');
                    }

                    if (result.data.hoststatus.hasOwnProperty('hostname') > 0 && $scope.hostNotFound === true) {
                        $scope.hostNotFound = false;
                        $('#host-not-found-modal-dialog').addClass('animated hinge');
                        $('#host-not-found-modal').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function () {
                            $('#host-not-found-modal').modal('hide');
                            $('#host-not-found-modal-dialog').removeClass('animated hinge');
                        });
                    }

                    if (result.data.hoststatus.hasOwnProperty('problem_has_been_acknowledged')) {
                        if (result.data.hoststatus.problem_has_been_acknowledged) {
                            $scope.loadAcknowledgementData();
                        } else {
                            $scope.acknowledgementData = null;
                        }
                    }

                    if (result.data.hoststatus.hasOwnProperty('scheduled_downtime_depth')) {
                        if (result.data.hoststatus.scheduled_downtime_depth > 0) {
                            $scope.loadHostDowntimeData();
                        } else {
                            $scope.downtimeData = null;
                        }
                    }

                }
            );
        };

        $scope.loadAcknowledgementData = function () {
            $http.get("/api/index.php/hostacknowledgements", {
                params: {
                    hostname: $scope.nodename,
                    limit: 1,
                    order: 'entry_time',
                    direction: 'desc'
                }
            }).then(function (result) {
                if (result.data.length > 0) {
                    $scope.acknowledgementData = result.data[0];
                }

            });
        };

        $scope.loadHostDowntimeData = function () {
            $http.get("/api/index.php/hostdowntime", {
                params: {
                    hostname: $scope.nodename
                }
            }).then(function (result) {
                if (result.data.length > 0) {
                    $scope.downtimeData = result.data[0];
                }

            });
        };

        $scope.submitPassiveCheckResult = function () {
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitPassiveHostCheckResult.html',
                controller: 'SubmitPassiveHostCheckController'
            });

            modal.result.then(function (data) {
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.sendCustomHostNotification = function () {
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitCustomHostNotification.html',
                controller: 'SubmitCustomHostNotificationController'
            });

            modal.result.then(function (data) {
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.submitHostDowntime = function () {
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitHostDowntime.html',
                controller: 'SubmitHostDowntimeController'
            });

            modal.result.then(function (data) {
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.submitHostAcknowledgement = function () {
            var modal = $uibModal.open({
                templateUrl: 'templates/modals/submitHostAcknowledgement.html',
                controller: 'SubmitHostAcknowledgementController'
            });

            modal.result.then(function (data) {
                $scope.sendCommandWithArgs(data);
            });
        };

        $scope.submitDeleteDowntime = function (internal_downtime_id) {
            if ($scope.isAllowedToSubmitCommand === false) {
                return;
            }
            var data = {};
            data['command_name'] = 'DEL_HOST_DOWNTIME';
            data['downtime_id'] = internal_downtime_id;
            data['node_name'] = $scope.data.hoststatus.node_name;

            $http.get("/api/index.php/externalcommand_args", {
                params: data
            }).then(function (result) {
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

        $scope.sendCommandWithArgs = function (data) {
            if ($scope.isAllowedToSubmitCommand === false) {
                return;
            }
            data['hostname'] = $scope.nodename;
            data['node_name'] = $scope.data.hoststatus.node_name;

            $http.get("/api/index.php/externalcommand_args", {
                params: data
            }).then(function (result) {
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

        $scope.sendCommand = function (commandString) {
            if ($scope.isAllowedToSubmitCommand === false) {
                return;
            }
            /**
             * @see src/Generators/ExternalCommand.php for the commandId values
             */

            var commandId = null;
            switch (commandString) {
                case 'notifications':
                    commandId = 27; //Enable
                    if ($scope.data.hoststatus.notifications_enabled === true) {
                        commandId = 26; //Disable
                    }
                    break;

                case 'activeChecks':
                    commandId = 21; //Enable
                    if ($scope.data.hoststatus.active_checks_enabled === true) {
                        commandId = 20; //Disable
                    }
                    break;

                case 'passiveChecks':
                    commandId = 29; //Enable
                    if ($scope.data.hoststatus.passive_checks_enabled === true) {
                        commandId = 28; //Disable
                    }
                    break;

                case 'flappDetection':
                    commandId = 25; //Enable
                    if ($scope.data.hoststatus.flap_detection_enabled === true) {
                        commandId = 24; //Disable
                    }
                    break;

                case 'eventHandler':
                    commandId = 23; //Enable
                    if ($scope.data.hoststatus.event_handler_enabled === true) {
                        commandId = 22; //Disable
                    }
                    break;

                case 'reschedule':
                    commandId = 2;
                    break;

                case 'reschedule_all':
                    commandId = 3;
                    break;

                default:
                    console.error('Unknown command');
                    return;
                    break;
            }

            $http.get("/api/index.php/externalcommand", {
                params: {
                    hostname: $scope.nodename,
                    node_name: $scope.data.hoststatus.node_name,
                    command: commandId
                }
            }).then(function (result) {
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

        $scope.getLoginState = function () {
            $http.get("/api/index.php/loginstate", {
                params: {}
            }).then(function (result) {
                $scope.isAllowedToSubmitCommand = false;
                if (result.data.canAnonymousSubmitCommand === true || result.data.isLoggedIn === true) {
                    $scope.isAllowedToSubmitCommand = true;
                }
            });
        };

        //triggers reload on load and on search
        $scope.$watch('[state_filter, servicedescription__like]', function () {
            if($scope.init === true){
                return;
            }
            $scope.reload();
        }, true);

        ReloadService.setCallback($scope.reload);

        $scope.getLoginState();
        $scope.reload();

    });