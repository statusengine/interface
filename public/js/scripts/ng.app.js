angular.module('Statusengine', ['ui.router', 'infinite-scroll', 'duScroll', 'ui.bootstrap'])


    .factory("httpInterceptor", function($q, $rootScope, $timeout){
        $rootScope.ajax_server_error = '';
        $rootScope.ajax_server_error_message = '';
        var hasAjaxServerError = false;

        return {
            response: function(result){
                $timeout(function(){
                    //Let the weel spin for one seconds
                    $rootScope.loading = false;
                }, 1000);

                if($rootScope.ajaxError === true){
                    $rootScope.ajaxError = false;
                    $('#ajax-modal-dialog').addClass('animated hinge');
                    $('#ajax-modal').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function(){
                        $('#ajax-modal').modal('hide');
                        $('#ajax-modal-dialog').removeClass('animated hinge');
                    });
                }

                if(hasAjaxServerError === true){
                    hasAjaxServerError = false;
                    $('#server-error-modal-dialog').addClass('animated hinge');
                    $('#server-error-modal').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function(){
                        $('#server-error-modal').modal('hide');
                        $('#server-error-modal-dialog').removeClass('animated hinge');
                    });
                }


                return result || $.then(result)
            },
            request: function(response){
                $rootScope.loading = true;
                return response || $q.when(response);
            },
            responseError: function(rejection){
                console.log(rejection);
                $rootScope.loading = false;
                if(rejection.status === -1){
                    $rootScope.ajaxError = true;
                    $('#ajax-modal').modal('show');
                }

                if(rejection.status == 401){
                    window.location = '/login.html';
                }

                if(rejection.status >= 500){
                    hasAjaxServerError = true;
                    $('#server-error-modal').modal('show');
                    $rootScope.ajax_server_error = '[' + rejection.status + '] ' + rejection.statusText;
                    $rootScope.ajax_server_error_message = JSON.stringify(rejection.data, null, 2);
                }

                return $q.reject(rejection);
            }
        };
    })

    .filter('isNotEmpty', function(){
        return function(obj){
            if(typeof obj == 'undefined'){
                return false;
            }
            return !angular.equals(obj, {}) && !angular.equals(obj, []);
        }
    })

    .filter('configEnabled', function(){
        return function(value){
            if(value === true){
                return 'Enabled';
            }
            return 'Disabled';
        }
    })

    .filter('stateType', function(){
        return function(stateType){
            if(stateType == true){
                return 'Hard';
            }
            return 'Soft';
        }
    })

    .filter('serviceStatusNameByStatusCode', function(){
        return function(statuscode){
            if(statuscode == 0){
                return "Ok";
            }

            if(statuscode == 1){
                return "Warning"
            }

            if(statuscode == 2){
                return "Critical"
            }

            return "Unknown";
        }
    })

    .filter('bootstrapClassStatusCodeHost', function(){
        return function(statuscode){
            if(statuscode == 0){
                return "success";
            }

            if(statuscode == 1){
                return "danger"
            }

            return "primary";
        }
    })

    .filter('bootstrapClassStatusCodeService', function(){
        return function(statuscode){
            if(statuscode == 0){
                return "success";
            }

            if(statuscode == 1){
                return "warning"
            }

            if(statuscode == 2){
                return "danger"
            }

            return "primary";
        }
    })

    .filter('statusNameByStatusCodeHost', function(){
        return function(statuscode){
            if(statuscode == 0){
                return "Up";
            }

            if(statuscode == 1){
                return "Down"
            }

            return "Unreachable";
        }
    })

    .filter('statusBgByStatusCodeService', function(){
        return function(statuscode){
            if(statuscode == 0){
                return 'bg-green';
            }

            if(statuscode == 1){
                return 'bg-yellow';
            }

            if(statuscode == 2){
                return 'bg-red';
            }

            return 'bg-primary';
        }
    })

    .filter('yesOrNo', function(){
        return function(value){
            if(value === true || value === 1 || value === 'true'){
                return 'Yes';
            }
            return 'No';
        }
    })

    .filter('iconByStatusCodeService', function(){
        return function(statuscode){
            if(statuscode == 0){
                return 'fa-check-circle-o';
            }

            if(statuscode == 1){
                return 'fa-exclamation-triangle';
            }

            if(statuscode == 2){
                return 'fa-times';
            }

            return 'fa-question-circle-o';
        }
    })

    .filter('base64encode', function(){
        return function(str){
            return btoa(str);
        }
    })

    .filter('encodeURI', function(){
        return function(str){
            return encodeURI(str);
        }
    })


    .config(function($httpProvider){
        $httpProvider.interceptors.push("httpInterceptor");
        $httpProvider.defaults.cache = false;
        if(!$httpProvider.defaults.headers.get){
            $httpProvider.defaults.headers.get = {};
        }
        // disable IE ajax request caching
        $httpProvider.defaults.headers.get['If-Modified-Since'] = '0';

    })
    
    .config(function($urlRouterProvider, $stateProvider){
        $urlRouterProvider.otherwise("/dashboard");

        $stateProvider
            .state('dashboard', {
                url: '/dashboard',
                templateUrl: "templates/views/dashboard.html",
                controller: "DashboardController"
            })
            .state('logentries', {
                url: '/logentries',
                templateUrl: "templates/views/logentries.html",
                controller: "LogentryController"
            })
            .state('nodes', {
                url: '/nodes/:show_state',
                params: {
                    show_state: {
                        value: null,
                        squash: true
                    }
                },
                templateUrl: "templates/views/nodes.html",
                controller: "NodeController"
            })
            .state('services', {
                url: '/services/:show_state',
                params: {
                    show_state: {
                        value: null,
                        squash: true
                    }
                },
                templateUrl: "templates/views/services.html",
                controller: "ServiceController"
            })
            .state('nodedetails', {
                url: '/nodedetails/:nodename',
                templateUrl: "templates/views/nodedetails.html",
                controller: "NodeDetailsController"
            })
            .state('cluster', {
                url: '/cluster',
                templateUrl: "templates/views/cluster.html",
                controller: "ClusterController"
            })
            .state('servicedetails', {
                url: '/servicedetails/:nodename/:servicedescription',
                templateUrl: "templates/views/servicedetails.html",
                controller: "ServiceDetailsController"
            })
            .state('problems', {
                url: '/problems',
                templateUrl: "templates/views/problems.html",
                controller: "ProblemController"
            })
            .state('nodechecks', {
                url: '/nodechecks/:nodename',
                templateUrl: "templates/views/nodechecks.html",
                controller: "NodeChecksController"
            })
            .state('nodestatechanges', {
                url: '/nodestatechanges/:nodename',
                templateUrl: "templates/views/nodestatechanges.html",
                controller: "NodeStateChangesController"
            })
            .state('nodenotifications', {
                url: '/nodenotifications/:nodename',
                templateUrl: "templates/views/nodenotifications.html",
                controller: "NodeNotificationsController"
            })
            .state('nodeacknowledgements', {
                url: '/nodeacknowledgements/:nodename',
                templateUrl: "templates/views/nodeacknowledgements.html",
                controller: "NodeAcknowledgementsController"
            })
            .state('servicechecks', {
                url: '/servicechecks/:nodename/:servicedescription',
                templateUrl: "templates/views/servicechecks.html",
                controller: "ServiceChecksController"
            })
            .state('servicestatechanges', {
                url: '/servicestatechanges/:nodename/:servicedescription',
                templateUrl: "templates/views/servicestatechanges.html",
                controller: "ServiceStateChangesController"
            })
            .state('serviceacknowledgements', {
                url: '/serviceacknowledgements/:nodename/:servicedescription',
                templateUrl: "templates/views/serviceacknowledgements.html",
                controller: "ServiceAcknowledgementsController"
            })
            .state('servicenotifications', {
                url: '/servicenotifications/:nodename/:servicedescription',
                templateUrl: "templates/views/servicenotifications.html",
                controller: "ServiceNotificationsController"
            })

            .state('scheduleddowntimes', {
                url: '/scheduleddowntimes',
                templateUrl: "templates/views/scheduleddowntimes.html",
                controller: "ScheduleddowntimesController"
            })

            .state('acknowledgements', {
                url: '/acknowledgements',
                templateUrl: "templates/views/acknowledgements.html",
                controller: "AcknowledgementsController"
            });

    });
