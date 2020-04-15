angular.module('Statusengine')
    .controller('SubmitHostDowntimeController', function($scope, $uibModalInstance){

        $scope.type = "SCHEDULE_HOST_SVC_DOWNTIME";
        $scope.start = "0";
        $scope.end = "54000"; // Stat in 15 minutes, angular wants this as string
        $scope.enableDateMode = false;

        $scope.submit = function(){
            if($scope.enableDateMode === true){

                var startDate = new Date(
                    $scope.startDate.getFullYear(),
                    $scope.startDate.getMonth(),
                    $scope.startDate.getDate(),
                    $scope.startTime.getHours(),
                    $scope.startTime.getMinutes()
                );

                var start = parseInt((startDate.getTime() / 1000), 10);

                var endDate = new Date(
                    $scope.endDate.getFullYear(),
                    $scope.endDate.getMonth(),
                    $scope.endDate.getDate(),
                    $scope.endTime.getHours(),
                    $scope.endTime.getMinutes()
                );

                var end = parseInt((endDate.getTime() / 1000), 10);

            }else{
                var date = new Date();
                var start = parseInt($scope.start, 10);
                var end = parseInt($scope.end, 10);

                start = parseInt(date.getTime() / 1000, 10) + start;
                end = start + end;
            }

            $uibModalInstance.close({
                command_name: $scope.type,
                start: start,
                end: end,
                comment: $scope.comment || ""
            });
        };

        $scope.$watch('enableDateMode', function(){
            if($scope.enableDateMode === true){

                var now = new Date();

                endNow = now.getTime() + (60 * 15 * 1000); //Add 15 minutes to the end time
                var endTime = new Date(endNow);

                $scope.startDate = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), now.getMinutes());
                $scope.startTime = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), now.getMinutes(), 0, 0);
                $scope.endDate = new Date(endTime.getFullYear(), endTime.getMonth(), endTime.getDate(), endTime.getHours(), endTime.getMinutes());
                $scope.endTime = new Date(endTime.getFullYear(), endTime.getMonth(), endTime.getDate(), endTime.getHours(), endTime.getMinutes(), 0, 0);
            }else{
                //Restore default values
                $scope.start = "0";
                $scope.end = "54000"; // Stat in 15 minutes, angular wants this as string
            }
        });

        $scope.cancel = function(){
            $uibModalInstance.dismiss('cancel');
        }
    });