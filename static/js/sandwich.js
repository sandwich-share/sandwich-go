var app = angular.module('sandwich', ['infinite-scroll', '$strap.directives']);

app.controller('MainCtrl', function($scope, $http, $timeout) {
  $scope.peerFiles = [];
  $scope.searchFiles = [];
  $scope.isPeerSearch = false;
  $scope.isFileSearch = false;
  $scope.loading = false;
  $scope.gotAll = false;
  $scope.alerts = [];
  $scope.settings = {};
  $scope.version = '';
  var step = 100;
  var peerIP = '';
  var peerPort = '';
  var peerPath = '';

  var newAlert = function(type, title, content) {
    $scope.alerts.push({
      type: type,
      title: title,
      content: content
    });
    $timeout(function() {
      $scope.alerts.shift();
    }, 5000);
  };

  //Fetch the peers on page load and then every 15 seconds TODO: Websockets
  function fetchPeers() {
    $http.get('/peers').success(function(data) {
      $scope.peerList = data;
    });
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

  $scope.killServer = function() {
    if (confirm("Are you sure you want to shut down?")) {
      $http.get('/kill');
    }
  }

  $scope.downloadFile = function(ip, file, type) {
    $http.get('/download', {params: {ip: ip, file: file, type: type}});
  };

  $http.get('/settings').success(function(data) {
    $scope.settings.port = data['LocalServerPort'];
    $scope.settings.dir = data['SandwichDirName'];
    $scope.settings.openBrowser = !data['DontOpenBrowserOnStart'];
  });

  $http.get('/version').success(function(data) {
    $scope.version = data;
  });

  $scope.saveSettings = function() {
    console.log($scope.settings.port);
    $http.post('/settings', {}, {params: {
      localport: $scope.settings.port,
      dirname: $scope.settings.dir,
      openbrowser: $scope.settings.openBrowser
    }}).success(function(){
      newAlert('success', undefined, 'Settings Saved');
    });
  };
});
