
$(document).ready(function() {
		
	if(!("WebSocket" in window)){
		$('#chatLog, input, button, #examples').fadeOut("fast");	
		$('<p>Oh no, you need a browser that supports WebSockets. How about <a href="http://www.google.com/chrome">Google Chrome</a>?</p>').appendTo('#container');		
	}else{
		//The user has WebSockets
//  	var host = "ws://localhost:8000/socket/server/startDaemon.php";
//  	var host = "ws://students.fhstp.ac.at:8080/socket/server/server2.php";
 	    var host = "ws://192.168.1.107:8080/socket/server/server2.php";
  	
  	var socket = $.websocket(host,{
  	    open: function(e){
  	    
  	        log_message("event","Connection to socket server established. readyState: "+this.readyState);
  	        this.send("init","user");
  	    },
  	    close: function(e){log_message("event","Connection to socket server lost. readyState: "+this.readyState);},
  	    message:function(e){ },
        events: {
        //werden vom server geschickt
            msg:function(e){
            //generic message logger
                log_message("message","Received: "+e.data);
            },
            voted:function(e){
                $('#question').html('<h1>'+e.data+'</h1>');
            },
            asking:function(e){
                $('#question').html('<h2>'+e.question+'</h2>');
                $('#choices').html('');
                $.each(e.possibilities, function(index,value){                   
                    $('#choices').append('<a href="#" class="choice" data-p="'+index+'">'+value+'</a>');
                });                
            }
  	    }  	
  	}); //end websocket
  	       		  
       function log_message(type,msg){
          console.log(msg);
          $('#chatLog').append('<p class="'+type+'">'+msg+'</p>');
        }//End message()
        $('.choice').live('click',function(){
            $('#choices').html('');
            var choice = $(this).data('p');
//            var choicename = $(this).text();
//            socket.send("vote",{votenumber:choice,possibility:choicename});
              socket.send("vote",{votenumber:choice});
        });

	  }// end: the user has websockets
	
});
