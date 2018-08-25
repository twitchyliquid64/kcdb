var app = angular.module('rfboi', ['ui.materialize']);

app.controller('BodyController', ["$scope", "$rootScope", function ($scope, $rootScope) {
    $scope.page = "home";

    if (window.location.href.indexOf('#rcfilters') != -1)
      $scope.page = 'rc-filters';

    $scope.changePage = function(pageName){
        $scope.page = pageName;
        $rootScope.$broadcast('page-change', {page: pageName});
        setTimeout(function(){
          MathJax.Hub.Queue(["Typeset",MathJax.Hub]);
        }, 200);
    };
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


app.controller('RcFilterController', ["$scope", "$rootScope", function($scope, $rootScope) {
  function draw(t, vertical, horizontal) {
    $scope.nodes.addChild(t.input_node);
    $scope.nodes.lastChild.scale(1);
    $scope.nodes.lastChild.position = {x: 40, y: 70};
    $scope.nodes.addChild(horizontal);
    $scope.nodes.lastChild.scale(1);
    $scope.nodes.lastChild.position = {x: 120, y: 76.5};
    $scope.nodes.addChild(new $scope.paperSurface.Path.Line([55, 76.5], [95, 76.5]));

    $scope.nodes.addChild(vertical);
    $scope.nodes.lastChild.scale(1);
    $scope.nodes.lastChild.position = {x: 200, y: 130};
    $scope.nodes.addChild(new $scope.paperSurface.Path.Line([140, 76.5], [200, 76.5]));
    $scope.nodes.addChild(new $scope.paperSurface.Path.Line([200, 76.5], [200, 105]));

    $scope.nodes.addChild(t.gnd_node);
    $scope.nodes.lastChild.scale(1);
    $scope.nodes.lastChild.position = {x: 200, y: 190};
    $scope.nodes.addChild(new $scope.paperSurface.Path.Line([200, 155], [200, 177.5]));

    $scope.nodes.addChild(t.output_node);
    $scope.nodes.lastChild.scale(1);
    $scope.nodes.lastChild.position = {x: 280, y: 70};
    $scope.nodes.addChild(new $scope.paperSurface.Path.Line([200, 76.5], [265, 76.5]));
  }

  $scope.init = function(){
    $scope.paperSurface = new paper.PaperScope();
    $scope.paperSurface.setup(document.getElementById('rcFilter'));
    $scope.paperSurface.project.currentStyle.strokeColor = 'black';

    $scope.nodes = new $scope.paperSurface.Group();
    $scope.templateNodes = new $scope.paperSurface.Group();
    $scope.templateNodes.importSVG('netlistsvg/lib/analog.svg', {
      insert: false,
      onLoad: function(g) {
        draw(g.children, g.children.capacitor_vertical, g.children.resistor_horizontal);
      },
    })

		$scope.paperSurface.view.draw();
  };

  $scope.neatFrequency = function(f){
    if (f > Math.pow(10, 9))
      return (f/Math.pow(10, 9)).toPrecision(4) + " Ghz";
    if (f > Math.pow(10, 6))
      return (f/Math.pow(10, 6)).toPrecision(4) + " Mhz";
    if (f > Math.pow(10, 3))
      return (f/Math.pow(10, 3)).toPrecision(4) + " khz";

    return f + " hz";
  };

  $scope.$watchGroup(['C', 'R'], function (newValue, oldValue, scope) {
    console.log(newValue, oldValue, scope);
    $scope.f3db = null;
    $scope.R = parseFloat($scope.R)
    $scope.C = parseFloat($scope.C)
    if ($scope.R != 0 && $scope.C != 0) {
      $scope.f3db = 1 / (2 * Math.PI * $scope.R * $scope.C * Math.pow(10, -12));
    }

  });

  $scope.init();
}]);
