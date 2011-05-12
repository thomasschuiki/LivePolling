#!/php -q
<?php  /*  >php -q server.php  */

error_reporting(E_ALL);
set_time_limit(0);
ob_implicit_flush();

/**
 * callback function 
 * @param WebSocketUser $user Current user
 * @param string $msg Data from user sent
 * @param WebSocketServer $server Server object
 */
function process($user, $msg, $server){

    $decoded = json_decode($msg);

    switch($decoded->command){
        case 'init':
            if($decoded->data === 'admin'){
                $command['type']='login';
                $command['data']=false;
                
                $server->send($user->socket,json_encode($command));
            }else if($decoded->data === 'user')
            {    
                //if new user connects send him the most recent question            
                $question = $server->getQuestion();
                if($question){
                    $send['type'] = 'asking';
                    $send['question'] = $question->text;
                    $send['possibilities'] = $question->choices;
                    $server->send($user->socket,json_encode($send));
                }
            }
        break;
        case 'vote': 
            $return = count_vote($user, $decoded->data, $server);
            $return['type'] = 'voted';
            $server->send($user->socket,json_encode($return));
            unset($return);
     //send statistics data to admin
            foreach($server->getAdmins() as $admin)
            {
                if($admin)
                {
                    $return = $server->getQuestion();
                    $return->type='statistics';
                    $server->send($admin->socket,json_encode($return));
                }
            }

        break;
        
        case 'login':
            $return = check_login($user,$decoded->data,$server);
            $returnmsg['type']='msg';
            $returnmsg['data']=$return['msg'];
            
            $command['type']='login';
            $command['data']=$return['cmd'];

            $server->send($user->socket,json_encode($returnmsg));
            $server->send($user->socket,json_encode($command));
            
        break;
        
        case 'statistics':
            $return = $server->getQuestion();
            $return->type='statistics';
            $server->send($user->socket,json_encode($return));
        break;
        
        case 'question':
            $decoded->data->type = 'asking';
            $server->createQuestion($decoded->data->question,$decoded->data->possibilities);
            foreach($server->getUsers() as $client){
                $server->send($client->socket,json_encode($decoded->data));
            }
        break;
    
    }
}
function count_vote($user,$data, $server)
{

    $return = array();
       $question = $server->getQuestion();
        if ($user->data['vote'] !== $question->id) //user hasn't voted yet
        {

            $question->votes[$data->votenumber]++;
            
            $user->data['vote'] = $question->id;

            $user->data['ip'] = $user->ip;
                    
            $return['data'] = "Thanks for voting!";
            
        }else if($user->data['vote'] === $question->id) //user already voted
        {
            $return['data'] = "You already voted!";
        }

    return $return;
}

function check_login($user,$data,$server)
{
    $return = array();

        if(!isset($user->data['login'])) $user->data['login'] = null;
        
        if ($user->data['login']!==true) //admin isn't logged in
        {
            if($data->adminname=='admin' && $data->adminpass=='admin'){
                $user->data['login'] = true;
                $server->addAdmin($user);

                $return['msg'] = "you are now logged in";
                $return['cmd'] = true;
            }else{
              $return['msg'] = "the password or the username didn't match";              
              $return['cmd'] = false;
            }
            
        }else if($user->data['login']===true) //admin is logged in
        {
            $return['msg'] = "you are already logged in";
            $return['cmd'] = true;
        }

    return $return;
}


require_once 'WebSocketServer.class.php';
// new WebSocketServer( socket address, socket port, callback function )
$webSocket = new WebSocketServer("192.168.1.107", 8080, 'process');
$webSocket->run();
