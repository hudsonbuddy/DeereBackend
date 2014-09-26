'use strict';


// Declare app level module which depends on filters, and services
angular.module('myApp', [
  'ngRoute',
  'ngResource',
  'myApp.filters',
  'myApp.services',
  'myApp.directives',
  'myApp.controllers'
]).
config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/alerts', {templateUrl: 'partials/alerts_partial.html', controller: 'AlertsController'});
  $routeProvider.when('/login', {templateUrl: 'partials/login_partial.html', controller: 'LoginController'});
  $routeProvider.otherwise({redirectTo: '/login'});
}]);
