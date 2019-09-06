/*
 * Convex hull algorithm - Library (JavaScript)
 *
 * Copyright (c) 2018 Project Nayuki
 * https://www.nayuki.io/page/convex-hull-algorithm
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program (see COPYING.txt and COPYING.LESSER.txt).
 * If not, see <http://www.gnu.org/licenses/>.
 */

var convexhull = new function() {

	// Returns a new array of points representing the convex hull of
	// the given set of points. The convex hull excludes collinear points.
	// This algorithm runs in O(n log n) time.
	this.makeHull = function(points) {
		var newPoints = points.slice();
		newPoints.sort(this.POINT_COMPARATOR);
		return this.makeHullPresorted(newPoints);
	};


	// Returns the convex hull, assuming that each points[i] <= points[i + 1]. Runs in O(n) time.
	this.makeHullPresorted = function(points) {
		if (points.length <= 1)
			return points.slice();

		// Andrew's monotone chain algorithm. Positive y coordinates correspond to "up"
		// as per the mathematical convention, instead of "down" as per the computer
		// graphics convention. This doesn't affect the correctness of the result.

		var upperHull = [];
		for (var i = 0; i < points.length; i++) {
			var p = points[i];
			while (upperHull.length >= 2) {
				var q = upperHull[upperHull.length - 1];
				var r = upperHull[upperHull.length - 2];
				if ((q.x - r.x) * (p.y - r.y) >= (q.y - r.y) * (p.x - r.x))
					upperHull.pop();
				else
					break;
			}
			upperHull.push(p);
		}
		upperHull.pop();

		var lowerHull = [];
		for (var i = points.length - 1; i >= 0; i--) {
			var p = points[i];
			while (lowerHull.length >= 2) {
				var q = lowerHull[lowerHull.length - 1];
				var r = lowerHull[lowerHull.length - 2];
				if ((q.x - r.x) * (p.y - r.y) >= (q.y - r.y) * (p.x - r.x))
					lowerHull.pop();
				else
					break;
			}
			lowerHull.push(p);
		}
		lowerHull.pop();

		if (upperHull.length == 1 && lowerHull.length == 1 && upperHull[0].x == lowerHull[0].x && upperHull[0].y == lowerHull[0].y)
			return upperHull;
		else
			return upperHull.concat(lowerHull);
	};


	this.POINT_COMPARATOR = function(a, b) {
		if (a.x < b.x)
			return -1;
		else if (a.x > b.x)
			return +1;
		else if (a.y < b.y)
			return -1;
		else if (a.y > b.y)
			return +1;
		else
			return 0;
	};

};
