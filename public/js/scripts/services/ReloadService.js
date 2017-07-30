angular.module('Statusengine')
    .service('ReloadService', function ($interval, $rootScope) {
        var _callback = null;
        var autoReloadEnabled = true;
        var autoreloadTimer = null;

        var isDisabledTemporary = false;

        if (window.localStorage.getItem('autoReloadEnabled') == 'false') {
            autoReloadEnabled = false;
        }

        var callCallback = function () {
            if (_callback !== null) {
                _callback();
            }
        };

        var handleTimer = function (disableTemporary) {
            if (autoReloadEnabled === true) {
                if(!disableTemporary){
                    window.localStorage.removeItem('autoReloadEnabled');
                }
                if (autoreloadTimer === null) {
                    autoreloadTimer = $interval(callCallback, 10000);
                }
            } else {
                $interval.cancel(autoreloadTimer);
                autoreloadTimer = null;
                if(!disableTemporary) {
                    window.localStorage.setItem('autoReloadEnabled', false);
                }
            }
        };

        $rootScope.isAutoReloadEnabled = autoReloadEnabled;
        handleTimer(false);
        return {
            setCallback: function (callback) {
                _callback = callback;
            },
            triggerReload: callCallback,
            setAutoReloadEnabled: function (value) {
                autoReloadEnabled = value;
                $rootScope.isAutoReloadEnabled = autoReloadEnabled;
                handleTimer(false);
            },
            setAutoReloadEnabledTemporary: function(value){
                autoReloadEnabled = value;
                isDisabledTemporary = !value;
                $rootScope.isAutoReloadEnabled = autoReloadEnabled;
                handleTimer(true);
            },
            getAutoReloadEnabled: function () {
                return autoReloadEnabled;
            },
            enableAutoloadIfRequired: function(){
                if(isDisabledTemporary){
                    autoReloadEnabled = true;
                    handleTimer(false);
                }
            }
        }
    });