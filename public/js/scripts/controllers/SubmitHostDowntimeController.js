angular.module('Statusengine')
    .controller('SubmitHostDowntimeController', function ($scope, $uibModalInstance) {

        $scope.type = "SCHEDULE_HOST_SVC_DOWNTIME";
        $scope.start = "0";
        $scope.end = "54000"; // Stat in 15 minutes, angular wants this as string

        $scope.submit = function (){

            var date = new Date();
            var start = parseInt($scope.start, 10);
            var end = parseInt($scope.end, 10);

            start = parseInt(date.getTime() / 1000, 10) + start;
            end = start + end;

            $uibModalInstance.close({
                command_name: $scope.type,
                start: start,
                end: end,
                comment: $scope.comment || ""
            });
        };

        $scope.cancel = function(){
            $uibModalInstance.dismiss('cancel');
        }
    });