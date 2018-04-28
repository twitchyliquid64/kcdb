
app.controller('ViewController', ["$scope", "$rootScope", "$http", function ($scope, $rootScope, $http) {
  $scope.loading = false;
  $scope.last_modified = null;
  $scope.module = {};

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
      url: '/module/details',
    }).then(function successCallback(response) {
      $scope.module = response.data
      $scope.loading = false;
    }, function errorCallback(response) {
      $scope.loading = false;
      $scope.error = response;
    });
  }
  $scope.redraw = paint;

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

    // lines
    if ($scope.module && $scope.module.lines)
      for (var i = 0; i < $scope.module.lines.length; i++) {
        var lObj = $scope.module.lines[i];
        var l = new $scope.paperSurface.Path.Line(lObj.start, lObj.end);
        l.strokeColor = l.fillColor = resolveColor('line', lObj.layer);
        l.strokeWidth = lObj.width * 15;
        $scope.componentLayer.addChild(l);
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
            pGroup.addChild(new $scope.paperSurface.Shape.Ellipse({
              center: pObj.position,
              size: new $scope.paperSurface.Point(pObj.size).divide(2),
            }));
            break;
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

    // text
    if ($scope.module && $scope.module.texts)
      for (var i = 0; i < $scope.module.texts.length; i++) {
        var tObj = $scope.module.texts[i];
        var tGroup = new $scope.paperSurface.Group();
        var pt = new $scope.paperSurface.PointText({
          point: tObj.position,
          content: '' + tObj.value,
          fillColor: resolveColor('text', tObj.layer),
          fontFamily: 'Lucida Console',
          fontSize: 1,
          justification: 'center',
        });
        pt.translate([0, tObj.position.y - pt.position.y]);
        pt.scale(tObj.size.x, tObj.size.y);
        tGroup.addChild(pt);
        $scope.componentLayer.addChild(tGroup);
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

  $scope.load();
}]);
