angular.module('Statusengine')
    .service('NightModeService', function(){
        var nightModeEnabled = false;

        if(window.localStorage.getItem('nightMode') === 'true'){
            nightModeEnabled = true;
        }

        return {
            isNightModeEnabled: function(){
                return nightModeEnabled;
            },
            enableNightMode: function(){
                nightModeEnabled = true;
                window.localStorage.setItem('nightMode', true);
            },
            disableNightMode: function(){
                nightModeEnabled = false;
                window.localStorage.removeItem('nightMode');
            }
        }
    });