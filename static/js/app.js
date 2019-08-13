var app = angular.module('kcdb', ['ui.materialize', 'angularMoment']);

app.controller('BodyController', ["$scope", "$rootScope", "$location", function ($scope, $rootScope, $location) {
    $scope.page = "search";

    $scope.changePage = function(pageName){
        $scope.page = pageName;
        $location.hash(pageName);
        $rootScope.$broadcast('page-change', {page: pageName});
    };

    switch ($location.hash()) {
    case 'sources':
      setTimeout(function(){
        $scope.$apply(function() {
          $scope.changePage('sources');
        });
      }, 20);
      break;
    }
}]);

app.filter('escape', function() {
  return window.encodeURIComponent;
});

function parseLocation(location) {
    var pairs = location.substring(1).split("&");
    var obj = {};
    var pair;
    var i;

    for (i in pairs) {
        if (pairs[i] === "")
            continue;

        pair = pairs[i].split("=");
        obj[decodeURIComponent(pair[0])] = decodeURIComponent(pair[1]);
    }

    return obj;
}


app.controller('SourcesController', ["$scope", "$http", "$rootScope", "$interval", function ($scope, $http, $rootScope, $interval) {
    $scope.loading = false;
    $scope.sources = [];
    $scope.ingest_status = null;

    $scope.load = function(){
      $scope.loading = true;
      $http({
        method: 'GET',
        url: '/sources/all',
      }).then(function successCallback(response) {
        $scope.sources = response.data
        $scope.loadStatus();
      }, function errorCallback(response) {
        $scope.loading = false;
        $scope.error = response;
      });
    }

    $scope.loadStatus = function(){
      $scope.loading = true;
      $http({
        method: 'GET',
        url: '/ingestor/status',
      }).then(function successCallback(response) {
        $scope.ingest_status = response.data
        $scope.loading = false;
      }, function errorCallback(response) {
        $scope.loading = false;
        $scope.error = response;
      });
    }

    $scope.hasUp = function(dateStr) {
      return !dateStr.startsWith('1970-');
    }
    $scope.ingestScheduledNow = function(ds) {
      return moment().isAfter(ds);
    }
    $scope.isNext = function(uid){
      return $scope.ingest_status && $scope.ingest_status.next_sources && $scope.ingest_status.next_sources[0].uid == uid;
    }


    $rootScope.$on('page-change', function(event, args) {
      if (args.page == 'sources'){
        if (!$scope.ingest_status) {
          $scope.updater = $interval($scope.loadStatus, 23 * 1000);
        }
        $scope.load();
      } else {
        if ($scope.updater) {
          $interval.cancel($scope.updater);
          $scope.updater =  null;
        }
      }
    });
}]);


app.controller('SearchController', ["$scope", "$http", "$rootScope", "$interval", "$window", function ($scope, $http, $rootScope, $interval, $window) {
    $scope.loading = false;
    $scope.symbolSearch = false;
    $scope.results = [];
    $scope.sources = {};
    $scope.queryFromURL = parseLocation($window.location.search)['query'];
    $scope.searchQ = '';
    $scope.hasSearched = false;
    if ($scope.queryFromURL) {
      $scope.searchQ = $scope.queryFromURL.replace(/\+/g, ' ');
      $scope.hasSearched = true;
      document.getElementById("searchInput").focus();
    }

    $scope.search = function(query){
      $scope.hasSearched = true;
      $scope.loading = true;
      $scope.error = null;
      $http({
        method: 'POST',
        url: '/search/all',
        data: {query: $scope.searchQ, symbolsOnly: $scope.symbolSearch},
      }).then(function successCallback(response) {
        $scope.results = response.data;
        $scope.loading = false;
      }, function errorCallback(response) {
        $scope.loading = false;
        $scope.error = response;
      });
    }

    $scope.showTag = function(source_uid) {
      return !!$scope.sources[source_uid].tag;
    }

    $scope.typeToggled = function(){
      $scope.results = [];
      if ($scope.symbolSearch)
        Materialize.toast('The Symbols database is in Alpha, so its pretty rough. Please report any issues :)', 6500);
    }

    $scope.loadSources = function(){
      $scope.loading = true;
      return $http({
        method: 'GET',
        url: '/sources/all',
      }).then(function successCallback(response) {
        for (var i = 0; i < response.data.length; i++) {
          $scope.sources[response.data[i].uid] = response.data[i];
        }
        $scope.loading = false;
      }, function errorCallback(response) {
        $scope.loading = false;
        $scope.error = response;
      });
    }
    $scope.loadSources().then(function(){
      if ($scope.queryFromURL)
        $scope.search($scope.searchQ);
    });

    // error info helpers.
    $scope.ec = function(){
      if (!$scope.error)return null;
      if ($scope.error.success === false)
        return 'N/A';
      return $scope.error.status;
    }
    $scope.exp = function(){
      if (!$scope.error)return null;
      if ($scope.error.status === -1)
        return "Network Error or server offline";
      if ($scope.error.success === false)
        return 'The server encountered a problem handling the request';
      return $scope.error.statusText;
    }
}]);

app.controller('SymbolViewController', ["$scope", "$rootScope", "$http", "$window", function ($scope, $rootScope, $http, $window) {
  $scope.loading = false;
  $scope.last_modified = null;
  $scope.symbol = {};
  $scope.path = window.location.pathname.substring('/symbol/'.length);
  $scope.query = parseLocation($window.location.search)['query'];

  $scope.load = function(user){
    $scope.loading = true;
    $http({
      method: 'GET',
      url: '/sym/details/' + $scope.path,
    }).then(function successCallback(response) {
      $scope.symbol = response.data[0];
      $scope.loading = false;
    }, function errorCallback(response) {
      $scope.loading = false;
      $scope.error = response;
    });
  }

  $scope.load();
}]);
