
app.controller('ViewController', ["$scope", "$rootScope", "$http", "$window", function ($scope, $rootScope, $http, $window) {
  $scope.loading = false;
  $scope.last_modified = null;
  $scope.module = {};
  $scope.path = window.location.pathname.substring('/footprint/'.length);
  $scope.query = parseLocation($window.location.search)['query'];

  $scope.canvas = document.getElementById('partsCanvas');
  $scope.canvas.style.width ='100%';
  $scope.canvas.addEventListener('mousewheel', canvasMousewheelEvent, false);
  $scope.paperSurface = new paper.PaperScope();
  $scope.paperSurface.setup($scope.canvas);
  $scope.paperSurface.settings.insertItems = false;
  _initLayers();
  $scope.paperSurface.view.onMouseDown = canvasMouseDownEvent;
  $scope.paperSurface.view.onMouseDrag = canvasMouseDragEvent;
  //$scope.paperSurface.view.onKeyDown = canvasKeyDownEvent;

  $scope.load = function(user){
    $scope.loading = true;
    $http({
      method: 'GET',
      url: '/module/details/' + $scope.path,
    }).then(function successCallback(response) {
      $scope.module = response.data
      $scope.loading = false;
    }, function errorCallback(response) {
      $scope.loading = false;
      $scope.error = response;
    });
  }
  $scope.redraw = paint;

  $scope.goto = function(){
    window.location = 'https://' + $scope.path.replace('::', '/tree/master/');
  }

  $scope.$watchGroup(['module'], function (newValue, oldValue, scope) {
    $scope.last_modified = moment(1000*parseInt($scope.module.tedit, 16));
    console.log("Module:", $scope.module);
    paint();
  });

  function resolveColor(type, layer) {
    if (type == 'drill') {
      return '#252525';
    }

    switch (layer) {
      case 'F.SilkS':
        return '#008484';
      case 'F.Cu':
        return '#840000';
      case 'F.Fab':
        return '#C2C200';
      case 'F.CrtYd':
        return '#484848';
      case '*.Cu':
        return '#847415';
    }
    return 'black';
  }

  function _initLayers(){
    $scope.componentLayer = new $scope.paperSurface.Layer();
    $scope.paperSurface.project.addLayer($scope.componentLayer);
  }
  function paint() {
    $scope.paperSurface.view.center = new $scope.paperSurface.Point(0, 0);
    $scope.componentLayer.removeChildren();

    if ($scope.module.graphics) {
      $scope.unsupported = undefined;
      for (var i = 0; i < $scope.module.graphics.length; i++) {
        var g = $scope.module.graphics[i];
        switch (g.type) {
          case 'fp_line':
            var l = new $scope.paperSurface.Path.Line(g.renderable.start, g.renderable.end);
            l.strokeColor = l.fillColor = resolveColor('line', g.renderable.layer);
            l.strokeWidth = g.renderable.width * 15;
            $scope.componentLayer.addChild(l);
            break;

          case 'fp_circle':
            var c = new $scope.paperSurface.Shape.Circle({
              center: g.renderable.center,
              radius: new $scope.paperSurface.Point(g.renderable.center).getDistance(g.renderable.end),
            });
            c.strokeColor = resolveColor('circle', g.renderable.layer);
            c.strokeWidth = g.renderable.width * 5;
            $scope.componentLayer.addChild(c);
            break;

          case 'fp_poly':
            var p = new $scope.paperSurface.Path(g.renderable.points);
            p.strokeColor = p.fillColor = resolveColor('polygon', g.renderable.layer);
            p.strokeWidth = g.renderable.width * 5;
            $scope.componentLayer.addChild(p);
            break;

          case 'fp_text':
            var tGroup = new $scope.paperSurface.Group();
            var pt = new $scope.paperSurface.PointText({
              point: g.renderable.position,
              content: '' + g.renderable.value,
              fillColor: resolveColor('text', g.renderable.layer),
              fontFamily: 'Lucida Console',
              fontSize: 1,
              justification: 'center',
            });
            pt.translate([0, g.renderable.position.y - pt.position.y]);
            pt.scale(g.renderable.effects.size.x, g.renderable.effects.size.y);
            tGroup.addChild(pt);
            $scope.componentLayer.addChild(tGroup);
            break;
          default:
          if (!$scope.unsupported) {
            $scope.unsupported = {};
          }
            $scope.unsupported[g.type] = g.type;
        }
      }
    }

    // pads
    if ($scope.module && $scope.module.pads)
      for (var i = 0; i < $scope.module.pads.length; i++) {
        var pObj = $scope.module.pads[i];
        var pGroup = new $scope.paperSurface.Group();
        switch (pObj.shape) {
          case 'rect':
            pGroup.addChild(new $scope.paperSurface.Shape.Rectangle({
              center: pObj.position,
              size: new $scope.paperSurface.Point(pObj.size).divide(2),
            }));
            break;
          case 'oval':
          case 'circle':
            pGroup.addChild(new $scope.paperSurface.Shape.Ellipse({
              center: pObj.position,
              size: new $scope.paperSurface.Point(pObj.size).divide(2),
            }));
            break;
        }
        if (pObj.position.z_present) {
          pGroup.rotate(pObj.position.z);
        }
        pGroup.strokeColor = pGroup.fillColor = resolveColor('pad', pObj.layers[0]);

        // TODO: support other kinds of drill holes.
        if (pObj.drill && pObj.drill.scalar > 0) {
          pGroup.addChild(new $scope.paperSurface.Shape.Circle({
            center: pObj.position,
            radius: pObj.drill.scalar/2.0,
            fillColor: resolveColor('drill', pObj.layers[0]),
          }));
        }

        var pt = new $scope.paperSurface.PointText({
          point: pObj.position,
          content: '' + pObj.pin,
          fillColor: 'white',
          fontFamily: 'Courier New',
          fontSize: 2,
          justification: 'center',
        });
        pt.translate([0, pObj.position.y - pt.position.y]);
        pGroup.addChild(pt);
        $scope.componentLayer.addChild(pGroup);
      }

    $scope.componentLayer.scale(10);
  }

  function canvasMouseDownEvent(event){
    $scope.lastPoint = $scope.paperSurface.view.projectToView(event.point);
    // var hitResult = $scope.paperSurface.project.hitTest(event.point, hitOptions);
    // setSelected(hitResult ? hitResult.item.name : '');
  }
  function canvasMouseDragEvent(event){
    var point = $scope.paperSurface.view.projectToView(event.point);
    var last = $scope.paperSurface.view.viewToProject($scope.lastPoint);
    $scope.paperSurface.view.scrollBy(last.subtract(event.point));
    $scope.lastPoint = point;
  }
  function canvasMousewheelEvent(event){
    $scope.paperSurface.view.scale(1 + (-0.0009 * event.deltaY));
    //$scope.paperSurface.view.center = $scope.paperSurface.view.viewToProject(new $scope.paperSurface.Point(event.layerX, event.layerY));
    return false;
  }






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



  $scope.load();
}]);
