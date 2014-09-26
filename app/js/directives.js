'use strict';

/* Directives */


angular.module('myApp.directives', [])
    .directive('appVersion', ['version', function(version) {
	return function(scope, elm, attrs) {
	    elm.text(version);
	};
    }])
    .directive("autoScroll",function(){
	return {
	    link : function(scope,elem,attr){

		var raw = elem[0];
		var funCheckBounds = function(evt) {
		    console.log("event fired: " + evt.type);
		    var rectObject = raw.getBoundingClientRect();
		    if (rectObject.bottom === window.innerHeight) {
			scope.$apply(attr.autoScroll);
		    }
		};
		
		angular.element(window).bind('scroll load', funCheckBounds);
	    }
	};
    });

