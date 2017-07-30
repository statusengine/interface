angular.module('Login', [])
    .factory("httpInterceptor", function ($q, $rootScope, $timeout) {
        $rootScope.ajax_server_error = '';
        $rootScope.ajax_server_error_message = '';
        var hasAjaxServerError = false;

        return {
            response: function (result) {
                if ($rootScope.ajaxError === true) {
                    $rootScope.ajaxError = false;
                    $('#ajax-modal-dialog').addClass('animated hinge');
                    $('#ajax-modal').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function () {
                        $('#ajax-modal').modal('hide');
                        $('#ajax-modal-dialog').removeClass('animated hinge');
                    });
                }

                if (hasAjaxServerError === true) {
                    hasAjaxServerError = false;
                    $('#server-error-modal-dialog').addClass('animated hinge');
                    $('#server-error-modal').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function () {
                        $('#server-error-modal').modal('hide');
                        $('#server-error-modal-dialog').removeClass('animated hinge');
                    });
                }


                return result || $.then(result)
            },
            request: function (response) {
                return response || $q.when(response);
            },
            responseError: function (rejection) {
                if (rejection.status === -1) {
                    $rootScope.ajaxError = true;
                    $('#ajax-modal').modal('show');
                }

                if (rejection.status >= 500) {
                    hasAjaxServerError = true;
                    $('#server-error-modal').modal('show');
                    $rootScope.ajax_server_error = '[' + rejection.status + '] ' + rejection.statusText;
                    $rootScope.ajax_server_error_message = JSON.stringify(rejection.data, null, 2);
                }

                return $q.reject(rejection);
            }
        };
    })

    .config(function ($httpProvider) {
        $httpProvider.interceptors.push("httpInterceptor");
    });