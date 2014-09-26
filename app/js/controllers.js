/* Controllers */

angular.module('myApp.controllers', ['ngSanitize']).controller('AlertsController', ['$scope', "Alerts", "Logout",function($scope,Alerts,Logout) {

    $scope.loadingMore = false;

    $scope.alertArray = [];

    $scope.topRightData = "";
    $scope.topRightHelper = false;

    $scope.toggleDateLoc = function(index){
        console.log("double click!");
        console.log(index);
        if($scope.alertArray[index].topRightHelper%2 == 0){

            $scope.alertArray[index].topRightData = $scope.alertArray[index].location.String;
            $scope.alertArray[index].topRightHelper++;
            $scope.topRightHelper = !$scope.topRightHelper;
           
        }else {

            $scope.alertArray[index].topRightData = $scope.alertArray[index].date.Date;
            $scope.alertArray[index].topRightHelper++;
            $scope.topRightHelper = !$scope.topRightHelper;
        };

    };

    $scope.loadMore = function(){
	if(!$scope.loadingMore){
	    $scope.loadingMore = true;
	    
	    Alerts.query({offset : $scope.alertArray.length},
			 function(alerts){
			     alerts.forEach(function(alertContainer){
                     alert = alertContainer.alert_data
                     if (["DF_MARGIN1", "DF_MARGIN2", "COV"].indexOf(alert.type) != -1){

                         alertContainer.show_data = {
                             Type : alert.type,
                             Current : alert.currentString,
                             Average : alert.aveString,
                             Min : alert.minString,
                             Max : alert.maxString,			     
                         };


                     }else if(["RIDEQUAL", "SINGULATION"].indexOf(alert.type) != -1){

                         alertContainer.show_data = {
                             Type : alert.type,
                             Current : alert.currentIntString + "%",
                             Average : alert.aveIntString + "%",
                             Min : alert.minIntString + "%",
                             Max : alert.maxIntString + "%",			     
                         };
                        

                     }else if(alert.type === "ACT_POP"){

                        alertContainer.show_data = {
                             Type : alert.type,
                             Current : (alert.current/1000).toFixed(2) + "k",
                             Average : (alert.ave/1000).toFixed(2) + "k",
                             Min : (alert.min/1000).toFixed(2) + "k",
                             Max : (alert.max/1000).toFixed(2) + "k",			     
                         };
                        

                     }

                     alert.ts = moment(alert.ts).fromNow();
                     console.log(alert.ts);

                     alertContainer.topRightData = alert.ts;
                     alertContainer.topRightHelper = 0;

                     alertContainer.date = {
                         Date : alert.ts,
                         Millis : alert.ts_millis,
                     };
                     if(alert.locationData){
                         alertContainer.location = {
                                 Lat : alert.locationData[0],
                                 Lng : alert.locationData[1],
                                 Heading : alert.locationData[2],
                                 String : "(" + alert.locationData[0] + ", " + alert.locationData[1] + ") " + alert.locationData[2] + "&deg;",
                         };
                     }else{
			 alertContainer.location = {Lat : "?", Lng : "?", Heading : "?", String: "??"};
		     }
                     console.log(alert.show_data);
                     });
			     $scope.alertArray= $scope.alertArray.concat(alerts);
			     $scope.loadingMore = false;
			 },
			 function(err){
			     location.href = "#/login";
			     $scope.loadingMore = false;
			 });
	}
    }
    
    $scope.doLogOut = function(){
	Logout.get({},
		   function(){
		       location.href = "#/login"
		   },
		   function(err){
		       location.href = "#/login"
		   }
		  );
    }

    $scope.loadMore();
    
}]).
    
controller('LoginController', ['$scope',"SessionStatus","Login","Logout", function($scope,SessionStatus,Login,Logout) {
    
    $scope.doLogin = function(){
	
	Login.save({
	    username : $scope.username,
	    password : $scope.password,
	},function(response){
	    location.href = "#/alerts";
	},function(err){
	    alert(err.data);
	    if(err.status != 401){
		Logout.get();
	    }
	});
	
	$scope.username = "";
	$scope.password = "";
    }
    SessionStatus.get({},
		      function(resp){
			  //aready logged in
			  location.href = "#/alerts";
		      },
		      function(err){
			  if(err.status == 401){
			      console.log("ready to log in");
			  }else{
			      alert("Server problems...");
			      Logout.get()
			  }
		      });
    
}]);
