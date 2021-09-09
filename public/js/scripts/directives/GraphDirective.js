angular.module('Statusengine').directive('graph', function ($http, $filter, $rootScope) {
    return {
        restrict: 'A',
        templateUrl: 'templates/directives/graph.html',
        scope: {
            'perfdataname': '=',
            'perfdatagauge': '=',
            'servicestatus': '=',
            'timespan': '=',
            'algorithm': '=',
            'displayperfdata': '='
        },
        controller: function ($scope) {
            var timespan = $scope.timespan;
            $scope.start = new Date().getTime() / 1000 | 0;
            $scope.end = new Date().getTime() / 1000 | 0;
            $scope.end = $scope.end - timespan;

            $scope.initializing = true;

            $scope.chartInstance = null;

            $scope.load = function () {
                $http.get("./api/index.php/serviceperfdata", {
                    params: {
                        hostname: $scope.servicestatus.hostname,
                        servicedescription: $scope.servicestatus.service_description,
                        metric: $scope.perfdataname,
                        limit: 500,
                        start: $scope.start,
                        end: $scope.end,
                        compression_algorithm: $scope.algorithm
                    }
                }).then(function (result) {
                    $scope.start = new Date().getTime() / 1000 | 0;
                    $scope.end = result.data.end;
                    $scope.perfdata = result.data.perfdata;
                });

            };

            $scope.$watch('servicestatus', function () {
                if (!$scope.servicestatus) {
                    return;
                }

                if(!$scope.displayperfdata){
                    return;
                }
                $scope.load();
            });

            $scope.$watchGroup(['timespan', 'algorithm'], function () {
                if(!$scope.initializing){

                    $scope.chartInstance.destroy();
                    $scope.chartInstance = null;

                    $scope.start = new Date().getTime() / 1000 | 0;
                    $scope.end = new Date().getTime() / 1000 | 0;
                    $scope.end = $scope.end - $scope.timespan;

                    $scope.load();
                }
                $scope.initializing = false;
            });

            $scope.getBackgroundColor = function(){
                if($scope.algorithm == 'min') {
                    return "rgba(0, 166, 90, 0.18)";
                }

                if($scope.algorithm == 'max') {
                    return "rgba(226, 77, 66, 0.18)";
                }

                return "rgba(31, 118, 189, 0.18)";
            },

            $scope.getBorderColor = function(){
                if($scope.algorithm == 'min') {
                    return "rgba(0, 200, 70, 1)";
                }

                if($scope.algorithm == 'max') {
                    return "rgba(255, 0, 0, 1)";
                }

                return "rgba(31, 120, 193, 1)";
            }

        },

        link: function ($scope, element) {
            $scope.$watch('perfdata', function () {
                if (!$scope.perfdata) {
                    return;
                }

                if ($scope.chartInstance === null) {
                    Chart.defaults.global.elements.point.radius = 1;
                    Chart.defaults.global.elements.point.hoverRadius = 4;
                    var chartValues = [];
                    var chartLabels = [];
                    for (var k in $scope.perfdata) {
                        var date = $filter('date')(($scope.perfdata[k].timestamp * 1000), "dd.MM.yy HH:mm");
                        chartLabels.push(date);
                        chartValues.push($scope.perfdata[k].value);
                    }

                    var ctx = $(element).find('canvas');
                    var label = $scope.perfdataname;

                    if ($scope.perfdatagauge.unit !== '' && $scope.perfdatagauge.unit !== null) {
                        label += ' in ' + $scope.perfdatagauge.unit;
                    }

                    $scope.chartInstance = new Chart(ctx, {
                        type: 'line',
                        data: {
                            labels: chartLabels,
                            datasets: [{
                                label: label,
                                data: chartValues,
                                backgroundColor: $scope.getBackgroundColor(),
                                borderColor: $scope.getBorderColor(),
                                borderWidth: 1
                            }]
                        },
                        options: {
                            responsive: true
                        }
                    });

                } else {
                    for (var k in $scope.perfdata) {
                        var date = $filter('date')(($scope.perfdata[k].timestamp * 1000), "dd.MM.yy HH:mm");
                        $scope.chartInstance.data.labels.shift();
                        $scope.chartInstance.data.labels.push(date);

                        $scope.chartInstance.data.datasets[0].data.shift();
                        $scope.chartInstance.data.datasets[0].data.push($scope.perfdata[k].value);

                    }

                    $scope.chartInstance.data.datasets[0].backgroundColor = $scope.getBackgroundColor();
                    $scope.chartInstance.data.datasets[0].borderColor = $scope.getBorderColor();

                    $scope.chartInstance.update();
                }

            });
        }
    };
});
