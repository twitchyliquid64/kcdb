
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
      var lastLine = null;
      var lastLinePoint = null;
      for (var i = 0; i < $scope.module.graphics.length; i++) {
        var g = $scope.module.graphics[i];
        switch (g.type) {
          case 'fp_line':
            var l = new $scope.paperSurface.Path.Line(g.renderable.start, g.renderable.end);
            l.strokeCap = 'round';
            l.strokeColor = resolveColor('line', g.renderable.layer);
            l.strokeWidth = g.renderable.width * 15;

            // Join it with the previous line if possible.
            if (lastLine && lastLinePoint.x == g.renderable.start.x && lastLinePoint.y == g.renderable.start.y) {
              lastLine.join(l);
              lastLinePoint = g.renderable.end;
              break;
            }
            $scope.componentLayer.addChild(l);
            lastLine = l;
            lastLinePoint = g.renderable.end;
            break;

          case 'fp_circle':
            drawCircle(g);
            break;

          case 'fp_arc':
            var startAngle = new $scope.paperSurface.Point(g.renderable.start).getAngle(g.renderable.end);

            var pts = [new $scope.paperSurface.Point(g.renderable.end)];
            for (var j = 0; j < 180; j++) {
              pts.push(new $scope.paperSurface.Point(g.renderable.end).rotate((g.renderable.angle*j/180.0) - startAngle, g.renderable.start));
            }
            var p = new $scope.paperSurface.Path(pts);
            p.strokeColor = resolveColor('circle', g.renderable.layer);
            p.strokeWidth = g.renderable.width * 15;
            $scope.componentLayer.addChild(p);
            break;

          case 'fp_poly':
            var p = new $scope.paperSurface.Path(g.renderable.points);
            p.strokeColor = p.fillColor = resolveColor('polygon', g.renderable.layer);
            p.strokeWidth = g.renderable.width * 5;
            $scope.componentLayer.addChild(p);
            break;

          case 'fp_text':
            drawText(g);
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
        renderBasePad(pObj, pGroup);
        $scope.componentLayer.addChild(pGroup);
      }

    $scope.componentLayer.scale(10);
  }

  function renderBasePad(pObj, pGroup) {
    var size = new $scope.paperSurface.Point(pObj.size).divide(2);
    var clearance = 0.2 / 2;
    if (pObj.clearance) {
      clearance = pObj.clearance / 2;
    }
    size.x -= clearance;
    size.y -= clearance;

    renderPadShape(pObj, pGroup, pObj.position, size, pObj.shape);
    pGroup.strokeColor = pGroup.fillColor = resolveColor('pad', pObj.layers[0]);

    // TODO: support other kinds of drill holes.
    if (pObj.drill_size.x > 0) {
      pGroup.addChild(new $scope.paperSurface.Shape.Circle({
        center: new $scope.paperSurface.Point(pObj.position).add(pObj.drill_offset),
        radius: pObj.drill_size.x/2,
        fillColor: resolveColor('drill', pObj.layers[0]),
      }));
    }

    var pt = new $scope.paperSurface.PointText({
      point: pObj.position,
      content: '' + pObj.pin,
      fillColor: 'white',
      fontFamily: 'Courier New',
      fontSize: 2 * 0.7,
      justification: 'center',
    });
    pt.translate([0, pObj.position.y - pt.position.y]);
    pGroup.addChild(pt);
    if (pObj.position.z_present) {
      pGroup.rotate(pObj.position.z);
    }
  }

  function renderPadShape(pObj, pGroup, pos, size, shape) {
    switch (shape) {
      case 'custom':
      // TODO: Work out issues with scaling the sizes.
      // switch (pObj.Options.anchor) {
      //   case 'rect':
      //     var r = new $scope.paperSurface.Shape.Rectangle(
      //       new $scope.paperSurface.Point(-size.x/2, -size.y/2),
      //       new $scope.paperSurface.Point(size.x/2, size.y/2),
      //     );
      //     r.translate(pObj.position);
      //     pGroup.addChild(r);
      //     break;
      // }

        for (var i = 0; i < pObj.Primitives.length; i++) {
          var prim = pObj.Primitives[i];
          switch (prim.type) {
            case 'gr_poly':
              var p = new $scope.paperSurface.Path(prim.renderable.points);
              p.translate(new $scope.paperSurface.Point(pos));
              p.strokeWidth = prim.renderable.width * 5;
              pGroup.addChild(p);
              break;

            case 'gr_circle':
              var c = new $scope.paperSurface.Shape.Circle({
                center: new $scope.paperSurface.Point(prim.renderable.center),
                radius: new $scope.paperSurface.Point(prim.renderable.center).getDistance(prim.renderable.end),
              });
              c.translate(pos);
              c.strokeWidth = prim.renderable.width * 5;
              pGroup.addChild(c);
              break;
          }
        }
        break;
      case 'rect':
        pGroup.addChild(new $scope.paperSurface.Shape.Rectangle({
          center: pObj.position,
          size: size,
        }));
        break;
      case 'roundrect':
        pGroup.addChild(new $scope.paperSurface.Shape.Rectangle({
          center: pObj.position,
          size: size,
          radius: Math.min(pObj.roundrect_rratio * Math.min(size.x, size.y), 0.25),
        }));
        break;
      case 'oval':
      case 'circle':
        pGroup.addChild(new $scope.paperSurface.Shape.Ellipse({
          center: pObj.position,
          size: size,
        }));
        break;
    }
  }

  function drawCircle(g) {
    var c = new $scope.paperSurface.Shape.Circle({
      center: g.renderable.center,
      radius: new $scope.paperSurface.Point(g.renderable.center).getDistance(g.renderable.end),
    });
    c.strokeColor = resolveColor('circle', g.renderable.layer);
    c.strokeWidth = g.renderable.width * 5;
    $scope.componentLayer.addChild(c);
  }

  function drawText(g) {
    var tGroup = new $scope.paperSurface.Group();
    var pt = new $scope.paperSurface.PointText({
      point: g.renderable.position,
      content: '' + g.renderable.value,
      fillColor: resolveColor('text', g.renderable.layer),
      fontFamily: 'Lucida Console',
      fontSize: 1,
      fontWeight: g.renderable.effects.bold ? 'bold' : '',
      justification: 'center',
    });
    pt.translate([0, g.renderable.position.y - pt.position.y]);
    pt.scale(g.renderable.effects.size.x * 0.7, g.renderable.effects.size.y * 0.7);
    tGroup.addChild(pt);
    $scope.componentLayer.addChild(tGroup);
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
