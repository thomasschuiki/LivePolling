<?php
/**
 * WebSocket extension class of phpWebSockets
 *
 * @author Moritz Wutz <moritzwutz@gmail.com>
 * @version 0.1
 * @package phpWebSockets
 */

class socketWebSocket extends socket
{
	private $clients = array();
	private $handshakes = array();

	public function __construct()
	{
		parent::__construct();

		$this->run();
	}

	/**
	 * Runs the while loop, wait for connections and handle them
	 */
	private function run()
	{
		while(true)
		{
			# because socket_select gets the sockets it should watch from $changed_sockets
			# and writes the changed sockets to that array we have to copy the allsocket array
			# to keep our connected sockets list
			$changed_sockets = $this->allsockets;

			# blocks execution until data is received from any socket
			$num_sockets = socket_select($changed_sockets,$write=NULL,$exceptions=NULL,NULL);

			# foreach changed socket...
			foreach( $changed_sockets as $socket )
			{
				# master socket changed means there is a new socket request
				if( $socket==$this->master )
				{
					# if accepting new socket fails
					if( ($client=socket_accept($this->master)) < 0 )
					{
						$this->console('socket_accept() failed: reason: ' . socket_strerror(socket_last_error($client)));
						continue;
					}
					// if it is successful push the client to the allsockets array
					else
					{
						$this->allsockets[] = $client;

						// using array key from allsockets array, is that ok?
						// i want to avoid the often array_search calls
						$socket_index = array_search($client,$this->allsockets);
						$this->clients[$socket_index] = new stdClass;
						$this->clients[$socket_index]->socket_id = $client;
						$this->handshakes[$socket_index] = false;

						$this->console($client . ' CONNECTED!');
					}
				}
				// client socket has sent data
				else
				{
					#$socket_index = array_search($socket,$this->allsockets);
				

					$bytes = @socket_recv($socket,$buffer,2048,0);
					// the client status changed, but theres no data ---> disconnect
					if( $bytes === 0 )
					{
						$this->disconnected($socket);
					}
					// there is data to be read
					else
					{
						//this is a new connection, no handshake yet
						if( !$this->handshakes[$socket_index] )
						{
							$this->do_handshake($buffer,$socket,$socket_index);   
						}
						// handshake already done, read data
						else
						{
						//not getting here

							$action = substr($buffer,1,$bytes-2); // remove chr(0) and chr(255)
							$this->console("<{$action}");

							if( method_exists('socketWebSocketTrigger',$action) )
							{
								$this->send($socket,socketWebSocketTrigger::$action());
							}
							
							else
							{
								for ($i = 0; $i <= 0; $i++) {
									$this->send($socket,"{$action}");
								}
							}
						}
					}
				}
			}
		}
	}

	/**
	 * Manage the handshake procedure
	 *
	 * @param string $buffer The received stream to init the handshake
	 * @param socket $socket The socket from which the data came
	 * @param int $socket_index The socket index in the allsockets array
	 */
	private function do_handshake($buffer,$socket,$socket_index)
	{
	
		$this->console('Requesting handshake... for '.$socket);

    list($resource,$host,$origin,$key1,$key2,$l8b) = $this->getheaders($buffer);
		$this->console('Handshaking...');

		$upgrade  = "HTTP/1.1 101 Web Socket Protocol Handshake\r\n" .
				        "Upgrade: WebSocket\r\n" .
				        "Connection: Upgrade\r\n" .
        //				"WebSocket-Origin: {$origin}\r\n" .
        //				"WebSocket-Location: ws://{$host}{$resource}\r\n\r\n" . 
                "Sec-WebSocket-Origin: " . $origin . "\r\n" .
                "Sec-WebSocket-Location: ws://" . $host . $resource . "\r\n" .  
                "\r\n" .
                $this->calcKey($key1,$key2,$l8b) . "\r\n".
                chr(0);
$this->console($upgrade);
		$this->handshakes[$socket_index] = true;

		socket_write($socket,$upgrade,strlen($upgrade));

		$this->console('Done handshaking...');
	}


	/**
	 * Extends the socket class send method to send WebSocket messages
	 *
	 * @param socket $client The socket to which we send data
	 * @param string $msg  The message we send
	 */
	protected function send($client,$msg)
	{
		$this->console(">{$msg}");

		parent::send($client,chr(0).$msg.chr(255));
	}

	/**
	 * Disconnects a socket an delete all related data
	 *
	 * @param socket $socket The socket to disconnect
	 */
	private function disconnected($socket)
	{
		$index = array_search($socket, $this->allsockets);
		if( $index >= 0 )
		{
			unset($this->allsockets[$index]);
			unset($this->clients[$index]);
			unset($this->handshakes[$index]);
		}

		socket_close($socket);
		$this->console($socket." disconnected!");
	}

	/**
	 * Parse the handshake header from the client
	 *
	 * @param string $req
	 * @return array resource,host,origin,key1,key2,l8b
	 *//**/
	private function getheaders($req)
	{
		$req  = substr($req,4); // RegEx kill babies 
		$res  = substr($req,0,strpos($req," HTTP"));
		$req  = substr($req,strpos($req,"Host:")+6);
		$host = substr($req,0,strpos($req,"\r\n"));
		$req  = substr($req,strpos($req,"Origin:")+8);
		$ori  = substr($req,0,strpos($req,"\r\n"));
		$req  = substr($req,strpos($req,"Sec-WebSocket-Key1:")+20);
		$key1 = substr($req,0,strpos($req,"\r\n"));
  	$req  = substr($req,strpos($req,"Sec-WebSocket-Key2:")+20);
		$key2 = substr($req,0,strpos($req,"\r\n"));
  	$l8b  = substr($req,strpos($req,-8));
		return array($res,$host,$ori,$key1,$key2,$l8b);
	}/**/
	/*
	private function getheaders($req){
    $r=$h=$o=null;
    if(preg_match("/GET (.*) HTTP/"               ,$req,$match)){ $r=$match[1]; }
    if(preg_match("/Host: (.*)\r\n/"              ,$req,$match)){ $h=$match[1]; }
    if(preg_match("/Origin: (.*)\r\n/"            ,$req,$match)){ $o=$match[1]; }
    if(preg_match("/Sec-WebSocket-Key1: (.*)\r\n/",$req,$match)){ $this->console("Sec Key1: ".$sk1=$match[1]); }
    if(preg_match("/Sec-WebSocket-Key2: (.*)\r\n/",$req,$match)){ $this->console("Sec Key2: ".$sk2=$match[1]); }
    if($match=substr($req,-8))                                                                  { $this->console("Last 8 bytes: ".$l8b=$match); }
    return array($r,$h,$o,$sk1,$sk2,$l8b);
  }/**/
  private function calcKey($key1,$key2,$l8b){
        //Get the numbers
        preg_match_all('/([\d]+)/', $key1, $key1_num);
        preg_match_all('/([\d]+)/', $key2, $key2_num);
        //Number crunching [/bad pun]
        $this->console("Key1: " . $key1_num = implode($key1_num[0]) );
        $this->console("Key2: " . $key2_num = implode($key2_num[0]) );
        //Count spaces
        preg_match_all('/([ ]+)/', $key1, $key1_spc);
        preg_match_all('/([ ]+)/', $key2, $key2_spc);
        //How many spaces did it find?
        $this->console("Key1 Spaces: " . $key1_spc = strlen(implode($key1_spc[0])) );
        $this->console("Key2 Spaces: " . $key2_spc = strlen(implode($key2_spc[0])) );
        if($key1_spc==0|$key2_spc==0){ $this->console("Invalid key");return; }
        //Some math
        $key1_sec = pack("N",$key1_num / $key1_spc); //Get the 32bit secret key, minus the other thing
        $key2_sec = pack("N",$key2_num / $key2_spc);
        //This needs checking, I'm not completely sure it should be a binary string
        return md5($key1_sec.$key2_sec.$l8b,1); //The result, I think
  }
	/**
	 * Extends the parent console method.
	 * For now we just set another type.
	 *
	 * @param string $msg
	 * @param string $type
	 */
	protected function console($msg,$type='WebSocket')
	{
		parent::console($msg,$type);
	}
}

?>
