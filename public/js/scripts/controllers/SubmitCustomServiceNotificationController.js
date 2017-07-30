angular.module('Statusengine')
    .controller('SubmitCustomServiceNotificationController', function ($scope, $uibModalInstance) {

        $scope.force = true;
        $scope.broadcast = false;

        $scope.submit = function (){
            $uibModalInstance.close({
                comment: $scope.comment || "",
                force: $scope.force,
                broadcast: $scope.broadcast,
                command_name: "SEND_CUSTOM_SVC_NOTIFICATION"
            });
        };

        $scope.cancel = function(){
            $uibModalInstance.dismiss('cancel');
        }
    });