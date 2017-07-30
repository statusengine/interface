angular.module('Statusengine')
    .controller('SubmitPassiveHostCheckController', function ($scope, $uibModalInstance) {

        $scope.state = '0';

        $scope.submit = function (){
            $uibModalInstance.close({
                state: $scope.state,
                output: $scope.output || "",
                command_name: "PROCESS_HOST_CHECK_RESULT"
            });
        };

        $scope.cancel = function(){
            $uibModalInstance.dismiss('cancel');
        }
    });