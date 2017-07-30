angular.module('Statusengine')
    .controller('SubmitServiceAcknowledgementController', function ($scope, $uibModalInstance) {

        $scope.sticky = true;

        $scope.submit = function (){
            $uibModalInstance.close({
                comment: $scope.comment || "",
                sticky: $scope.sticky,
                command_name: "ACKNOWLEDGE_SVC_PROBLEM"
            });
        };

        $scope.cancel = function(){
            $uibModalInstance.dismiss('cancel');
        }
    });