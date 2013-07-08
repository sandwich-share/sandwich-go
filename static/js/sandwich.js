var app = angular.module('sandwich', ['infinite-scroll']);

app.controller('MyCtrl', function($scope, $http, $timeout) {
  $scope.peerFiles = [];
  $scope.searchFiles = [];
  $scope.isPeerSearch = false;
  $scope.isFileSearch = false;
  $scope.loading = false;
  $scope.gotAll = false;
  var step = 100;
  var peerIP = '';
  var peerPort = '';
  var peerPath = '';

  //Fetch the peers on page load and then every 15 seconds TODO: Websockets
  function fetchPeers() {
    $http.get('/peers').success(function(data) {
      $scope.peerList = data;
    });
    $timeout(fetchPeers, 15000);
  }
  fetchPeers(); //Making it a self executing function isn't working

  $scope.peerUpPath = function() {
    return peerPath.replace(/\/?[^/]+$/,'');
  };

  $scope.fileUrl = function(fileName, ip, port) {
    ip = ip || peerIP;
    port = port || peerPort;
    return "http://" + ip + ":" + port + "/files/" + fileName;
  };

  $scope.fetchPeerFiles = function(path, ip, port) {
    $scope.loading = true;

    if (path) {
      peerPath = path;
      $scope.peerFiles = [];
      $scope.gotAll = false;
    }

    if (ip) {
      $scope.isPeerSearch = true;
      $scope.isFileSearch = false;
      peerPort = port;
      peerIP = ip;
    }

    $http.get('/peer', {params: {peer: peerIP, path: peerPath, start: $scope.peerFiles.length, step: step}}).success(function(data) {
      if (!data.length) {
        $scope.gotAll = true;
      }
      $scope.peerFiles = $scope.peerFiles.concat(data);
      $scope.loading = false;
    });
  };

  $scope.fetchSearchFiles = function(search, regex) {
    $scope.loading = true;
    if (search) {
      $scope.searchFiles = [];
      $scope.isPeerSearch = false;
      $scope.isFileSearch = true;
      $scope.search = '';
      $scope.gotAll = false;
    }

    $http.get('/search', {params: {search: search, regex: regex, start: $scope.searchFiles.length, step: step}}).success(function(data) {
      if (!data.length) {
        $scope.gotAll = true;
      }
      $scope.searchFiles = $scope.searchFiles.concat(data);
      $scope.loading = false;
    });
  };

  $scope.loadMore = function() {
    if ($scope.gotAll) return;
    if ($scope.isFileSearch) {
      $scope.fetchSearchFiles();
    } else if ($scope.isPeerSearch) {
      $scope.fetchPeerFiles();
    }
  };

  $scope.downloadFile = function(ip, file, type) {
    $http.get('/download', {params: {ip: ip, file: file, type: type}});
  };
});
