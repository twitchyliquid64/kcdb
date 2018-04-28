var app = angular.module('kcdb', ['ui.materialize', 'angularMoment']);

app.controller('BodyController', ["$scope", "$rootScope", function ($scope, $rootScope) {
    $scope.page = "search";
    $scope.changePage = function(pageName){
        $scope.page = pageName;
        $rootScope.$broadcast('page-change', {page: pageName});
    };
}]);

app.controller('SourcesController', ["$scope", "$http", "$rootScope", function ($scope, $http, $rootScope) {
    $scope.loading = false;
    $scope.sources = [];

    $scope.load = function(user){
      $scope.loading = true;
      $http({
        method: 'GET',
        url: '/sources/all',
      }).then(function successCallback(response) {
        $scope.sources = response.data
        $scope.loading = false;
      }, function errorCallback(response) {
        $scope.loading = false;
        $scope.error = response;
      });
    }


    $rootScope.$on('page-change', function(event, args) {
      if (args.page == 'sources'){
        $scope.load();
      }
    });
}]);
