angular.module('Statusengine').directive('externalUrls', function(){
    return {
        restrict: 'E',
        templateUrl: 'templates/directives/external_urls.html',
        scope: {
            'data': '=',
            'externalUrls': '='
        },
        controller: function($scope){

            $scope.links = [];

            var replaceMacros = function(urlToReplace){
                macros = [
                    'hostname',
                    'service_description',
                    'node_name',
                    'current_check_attempt',
                    'max_check_attempts',
                    'current_state',
                    'is_hardstate',
                    'output',
                    'scheduled_downtime_depth',
                    'problem_has_been_acknowledged',
                    'last_check',
                    'next_check',
                    'last_state_change'
                ];

                for(var i in macros){
                    var macro = macros[i];

                    var search = '$' + macro + '$';
                    var replace = 'undefined';

                    if($scope.data.hasOwnProperty(macro)){
                        replace = $scope.data[macro];
                    }

                    urlToReplace = urlToReplace.replace(search, encodeURIComponent(replace));
                }

                return urlToReplace;
            };

            // Fire on page load
            for(var i in $scope.externalUrls){
                $scope.links.push({
                    name: $scope.externalUrls[i].name,
                    url: replaceMacros($scope.externalUrls[i].url)
                })
            }
        }
    };
});