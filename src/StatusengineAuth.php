<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine;

use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;
use Statusengine\Backend\StorageBackend;
use Statusengine\Exceptions\UnknownAuthTypeException;
use Statusengine\ValueObjects\User;

/**
 * Class StatusengineAuth
 * @package Statusengine
 *
 * Many thanks to https://github.com/tuupola/slim-basic-auth
 * which is a good place to look how to implement a own
 * Slim Middelware :)
 */
class StatusengineAuth {

    /**
     * @var Config
     */
    private $Config;

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var string
     */
    private $path;

    /**
     * @var array
     */
    private $urlsWithoutLogin;

    /**
     * @var SessionHandler
     */
    private $Session;

    /**
     * StatusengineAuth constructor.
     * @param Config $Config
     */
    public function __construct(Config $Config, StorageBackend $StorageBackend) {
        $this->Config = $Config;
        $this->StorageBackend = $StorageBackend;
        $this->urlsWithoutLogin = $this->Config->getUrlsWithoutLogin();
        $this->Session = new SessionHandler();
    }


    /**
     * StatusengineAuth invokable method
     * Called by the Slim Framework Middleware handler
     *
     * @link https://www.slimframework.com/docs/concepts/middleware.html
     * @param RequestInterface $request
     * @param ResponseInterface $response
     * @param callable $next
     * @return mixed
     */
    public function __invoke(RequestInterface $request, ResponseInterface $response, callable $next) {

        $this->path = '/this/url/will/never/exists/and/is/forbidden'; //Just for paranoia

        $this->path = $request->getUri()->getPath();

        //$response = $response->write(json_encode($request->getParsedBody()));


        //Is login required or is annonymous browsing allowed?
        if ($this->Config->isAnonymousAllowed() === true) {
            //Call the next loaded middleware and return

            if ($request->isPost()) {
                //User may be want to login?
                if ($this->login($request, $response) === false) {
                    $response = $response->withStatus(401);
                    return $this->error($request, $response, [
                        'error' => 'Authentication failed'
                    ]);
                };
                $this->Session->writeClose();
            }

            return $next($request, $response);
        }

        //Check if the user is logged in
        if ($this->isAuthenticated() === false && $request->isPost()) {
            //If not, try to login the user if its a post request
            try {
                $this->login($request, $response);
                $this->Session->writeClose();
            } catch (\Exception $exception) {
                $response = $response->withStatus(500);
                return $this->error($request, $response, [
                    'error' => $exception->getMessage()
                ]);
            }
        }

        //Is the user now logged in? Or is anonymous browsing allowed?
        if ($this->isAuthenticated() === false && $this->Config->isAnonymousAllowed() === false) {
            //Is this a page, we can access without login?
            //This is checked at this point, because in_array is may be slow
            if ($this->isAllowedUrlWithoutLogin() === true && $request->isGet() === true) {
                //All good, page is accessible without login
                return $next($request, $response);
            }

            //User has no login -> kick him out!
            $response = $response->withStatus(401);
            return $this->error($request, $response, [
                'error' => 'Authentication failed'
            ]);
        }

        $this->Session->writeClose();

        //Call the next loaded middleware
        return $next($request, $response);
    }

    /**
     * @return bool
     */
    public function isAuthenticated() {
        if ($this->Session->has('loginSuccessfully')) {
            if ($this->Session->get('loginSuccessfully') === true) {
                return true;
            }
        }
        return false;
    }

    /**
     * @param RequestInterface $request
     * @param ResponseInterface $response
     * @return bool
     * @throws UnknownAuthTypeException
     */
    private function login(RequestInterface $request, ResponseInterface $response) {
        //Do not store the password in $_SESSION
        //$_SESSION is saved in plain text on the disk or memcached or $whatever
        //So we don't want to store passwords in plain text!
        //And even there is no use case to store the password in $_SESSION

        $params = $request->getParsedBody();

        if (!isset($params['username']) || !isset($params['password'])) {
            return false;
        }
        $username = $params['username'];
        $password = $params['password'];
        $authType = $this->Config->getAuthType();

        $this->Session->set('username', $username);

        switch ($authType) {
            case 'basic':
                $result = $this->basicAuth($username, $password);
                $this->Session->set('loginSuccessfully', $result);
                return $result;
                break;

            case 'ldap':
                $ldap = new Ldap($this->Config);
                $result = $ldap->auth($username, $password);
                $this->Session->set('loginSuccessfully', $result);
                return $result;
                break;

            default:
                throw new UnknownAuthTypeException(sprintf(
                    'auth_type %s is not supported yet - sorry',
                    $authType
                ));
                break;
        }


        return false;
    }

    /**
     * @return bool
     */
    private function isAllowedUrlWithoutLogin() {
        return in_array($this->path, $this->urlsWithoutLogin, true);
    }

    /**
     * @param RequestInterface $request
     * @param ResponseInterface $response
     * @param array $data
     */
    private function error(RequestInterface $request, ResponseInterface $response, $data) {
        return $response->write(json_encode($data));
    }

    /**
     * @param string $username
     * @param string $password
     * @return bool
     */
    private function basicAuth($username, $password) {
        $UserLoader = $this->StorageBackend->getUserLoader();
        $dbResult = $UserLoader->getUserByUsername($username);

        if ($dbResult) {
            $User = new User($username, $dbResult['password']);
            return $User->isPasswordValid($password);
        }
        return false;
    }

}
