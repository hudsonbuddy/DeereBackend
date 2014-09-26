'use strict';

/* Services */


// Demonstrate how to register services
// In this case it is a simple value service.
angular.module('myApp.services', ['ngResource'])
    .factory("Login",function($resource){
	return $resource("/api/login");
    })
    .factory("Alerts",function($resource){
	return $resource("/api/alerts");
    })
    .factory("SessionStatus",function($resource){
	return $resource("/api/session");
    })
    .factory("Logout",function($resource){
	return $resource("/api/logout");
    })
    .value('version', '0.1');
